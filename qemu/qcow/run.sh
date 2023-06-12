qemu-system-x86_64 \
    -m 2048 \
    -drive file=UbuntuServer.qcow2,format=qcow2,index=0,media=disk \
    -device e1000,netdev=tape0 \
    -netdev tap,id=tape0,ifname=tap0,script=no,downscript=no \
    -vnc :1