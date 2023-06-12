image=$1

if [ -z "$image" ]; then
    echo "Usage: $0 <image>"
    exit 1
fi

pth=$(pwd)

docker run -d -p 5555:5901 -v $pth/img/img:/tmp --privileged --name qemu -it --entrypoint /bin/bash --rm --network docker_gwbridge $image