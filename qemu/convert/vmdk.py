import os
import argparse

__QEMU_IMG_BIN__ = "/usr/bin/qemu-img"
__QEMU_IMG_MIRROR__ = "./vmdk"

# convert vmdk to qcow2
def qemu_img_convert(src, dst):
    cmd = "%s convert -f vmdk -O qcow2 %s %s" % (__QEMU_IMG_BIN__, src, dst)
    os.system(cmd)

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Convert vmdk to qcow2")
    parser.add_argument("-s", "--src", help="source vmdk file")
    parser.add_argument("-d", "--dst", help="destination qcow2 file")
    args = parser.parse_args()

    if args.src is None or args.dst is None:
        parser.print_help()
        exit(1)

    qemu_img_convert(args.src, args.dst)