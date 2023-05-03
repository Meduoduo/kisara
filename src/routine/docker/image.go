package docker

/*
	This file is used to manage docker image routine
	To be a useful docker manage tools
	It's nessessary to automaticlly pull image in need
	And if there has a long time a image has not been used
	It's also important to remove it
	But this can cause a problem
	What if a image has been pulled as soon as it has been removed?
	This should be considered in the future.
*/

import (
	"container/list"
	"errors"
	"io"
	"time"

	"github.com/Yeuoly/kisara/src/helper"
	db "github.com/Yeuoly/kisara/src/routine/db"
	log "github.com/Yeuoly/kisara/src/routine/log"
	kisara_types "github.com/Yeuoly/kisara/src/types"
	"github.com/docker/docker/api/types"
	"github.com/shirou/gopsutil/disk"
)

var image_mutex helper.HighGranularityMutex[string]

var last_update time.Time

const (
	IMAGE_EXPIRE_DURATION = time.Hour * 24 * 30
)

// lock a image, it will wait for all running operation of a image
// inlcuding launch container, delete image etc.

func (c *Docker) PullImage(image_name string, event_callback func(message string)) (*kisara_types.Image, error) {
	// check if disk space is enough
	disk_usage, err := disk.Usage("/")
	if err == nil {
		if disk_usage.UsedPercent > 80 {
			log.Info("[Docker] Disk usage is too high, try to clean up...")
			c.updateImage()
		}
	} else {
		log.Error("[Docker] Failed to get disk usage: " + err.Error())
	}

	image := &kisara_types.Image{
		Name: image_name,
	}

	reader, err := c.Client.ImagePull(*c.Ctx, image_name, types.ImagePullOptions{})

	if err != nil || reader == nil {
		return nil, err
	}

	for {
		buf := make([]byte, 1024)
		n, err := reader.Read(buf)

		if err == nil && event_callback != nil {
			event := string(buf[:n])
			event_callback(event)
		}

		if err == io.EOF || n == 0 {
			break
		}

		if err != nil {
			return nil, err
		}
	}

	raw_image, _, err := c.Client.ImageInspectWithRaw(*c.Ctx, image_name)
	if err != nil {
		return nil, err
	}

	image.Uuid = raw_image.ID

	image_db := &kisara_types.DBImage{
		ImageName: image_name,
		ImageId:   image.Uuid,
		LastUsage: time.Now(),
	}

	err = db.CreateGeneric(image_db)
	if err != nil {
		return nil, err
	}

	return image, nil
}

// when launch container, it's nessesscry to require image first
// it will lock the image to avoid image deletion before launch container
// it will also automatically pull image if not exists
func (c *Docker) RequireImage(image_name string, message_callback func(string)) (*kisara_types.Image, error) {
	image_raw, _, err := c.Client.ImageInspectWithRaw(*c.Ctx, image_name)
	if err != nil {
		return nil, err
	}

	image := &kisara_types.Image{}
	image_id := image_raw.ID

	if image_id == "" {
		message_callback("image not found, try pull...")
		image_pull, err := c.PullImage(image_name, func(message string) {
			message_callback("message from pull image... pulling...")
		})

		if err != nil {
			return nil, errors.New("image not found")
		}

		image = image_pull
		message_callback("image pull finished")
	} else {
		image.Uuid = image_raw.ID
		if len(image_raw.RepoTags) == 0 {
			return nil, errors.New("image format not allowed")
		}

		image.Name = image_raw.RepoTags[0]
	}

	image_mutex.Lock(image_id)
	// TODO: update image latest usage
	image_record, err := db.GetGenericOne[kisara_types.DBImage](
		db.GenericEqual("image_id", image_id),
	)

	if err == db.ErrNotFound {
		image_record = kisara_types.DBImage{
			ImageId:   image_id,
			ImageName: image.Name,
			LastUsage: time.Now(),
		}

		// create new image record
		err = db.CreateGeneric(&image_record)
		if err != nil {
			return nil, err
		}
	} else {
		// update image record
		image_record.LastUsage = time.Now()
		err = db.UpdateGeneric(&image_record)
		if err != nil {
			return nil, err
		}
	}

	image_mutex.Unlock(image_id)

	return image, nil
}

// init current image db records, create record if not exists
func (c *Docker) InitImage() error {
	images, err := c.ListImage()
	if err != nil {
		return err
	}

	if images == nil {
		return errors.New("init image failed")
	}

	for _, image := range *images {
		image_record, err := db.GetGenericOne[kisara_types.DBImage](
			db.GenericEqual("image_id", image.Uuid),
		)

		if err == db.ErrNotFound {
			image_record = kisara_types.DBImage{
				ImageId:   image.Uuid,
				ImageName: image.Name,
				LastUsage: time.Now(),
			}

			// create new image record
			err = db.CreateGeneric(&image_record)
			if err != nil {
				return err
			}

			log.Info("[InitImage] initialize db record for image %v successfully", image_record.ImageName)
		}
	}

	return nil
}

// check image status and delete the oldest image, ensure the image is at the newest status
func (c *Docker) updateImage() {
	if time.Since(last_update) < time.Hour*1 {
		return
	}
	last_update = time.Now()

	images, err := c.Client.ImageList(*c.Ctx, types.ImageListOptions{
		All: true,
	})

	if err != nil {
		log.Error("[UpdateImage] Failed to get image list: %v", err)
	}

	image_relation_tree := helper.Tree[string, types.ImageSummary]{}
	// to avoid push child before parent layer, use a queue
	image_push_queue := list.New()
	// build image relation tree
	for _, image := range images {
		image_push_queue.PushBack(&image)
	}

	for image_push_queue.Len() != 0 {
		e := image_push_queue.Front()
		image_push_queue.Remove(e)
		image := e.Value.(*types.ImageSummary)
		// no parent, push to tree directly
		if len(image.ParentID) == 0 {
			image_relation_tree.AddParent(image.ID, image)
		} else {
			// try push to parent node, if failed, push back to queue
			if image_relation_tree.AddToParent(image.ParentID, image.ID, image) != nil {
				image_push_queue.PushBack(image)
			}
		}
	}

	// walk from bottom to root
	image_relation_tree.WalkReverse(func(id string, v *types.ImageSummary) {
		// check if current expired
		image_mutex.Lock(id)
		if !c.imageHasExpired(id) {
			image_mutex.Unlock(id)
			return
		}
		image_mutex.Lock(id)

		// get all child nodes
		current := image_relation_tree.GetNode(id)
		if current == nil {
			return
		}

		children := current.GetAllChildren()
		if children == nil {
			return
		}

		// lock all images which depends on current image
		for k := range children {
			image_mutex.Lock(k)
		}

		defer func() {
			for k := range children {
				image_mutex.Unlock(k)
			}
		}()

		// check if current expired
		if !c.imageHasExpired(id) {
			return
		}

		// delete current image
		images, err := c.Client.ImageRemove(*c.Ctx, id, types.ImageRemoveOptions{
			Force:         true,
			PruneChildren: true,
		})

		if err != nil {
			log.Error("[UpdateImage] Failed to remove image: %v", err)
			return
		}

		for _, image := range images {
			log.Info("[UpdateImage] Image removed: %v:%v", image.Untagged, image.Deleted)
		}
	})
}

func (c *Docker) imageHasExpired(image_id string) bool {
	record, err := db.GetGenericOne[kisara_types.DBImage](
		db.GenericEqual("image_id", image_id),
	)

	if err != nil {
		return false
	}

	return record.IsExpired(IMAGE_EXPIRE_DURATION)
}
