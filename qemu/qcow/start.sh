#!/bin/bash

get_eth0_ip() {
    local eth0_ip=$(ip -o -f inet addr show eth0 | awk '{print $4}')
    local ip=$(echo $eth0_ip | cut -d'/' -f1)
    echo "$ip"
}

get_eth0_ip_with_mask() {
    local eth0_ip=$(ip -o -f inet addr show eth0 | awk '{print $4}')
    echo "$eth0_ip"
}

get_eth0_subnet_mask() {
    local mask=$(ifconfig eth0 | grep -oP 'netmask \K\S+')
    echo "$mask"
}

get_eth0_gateway() {
    local getway=$(ip route show dev eth0 | grep default | awk '{print $3}')
    echo "$getway"
}

eth0_ip_with_mask=$(get_eth0_ip_with_mask)
eth0_ip=$(get_eth0_ip)
eth0_subnet_mask=$(get_eth0_subnet_mask)
eth0_gateway=$(get_eth0_gateway)

echo "eth0_ip_with_mask: $eth0_ip_with_mask"
echo "eth0_ip: $eth0_ip"
echo "eth0_subnet_mask: $eth0_subnet_mask"
echo "eth0_gateway: $eth0_gateway"

ifconfig eth0 down 
brctl addbr br0 
brctl addif br0 eth0 
brctl stp br0 off 
brctl setfd br0 0
brctl sethello br0 1
ifconfig br0 0.0.0.0 promisc up 
ifconfig eth0 0.0.0.0 promisc up 
# set up the bridge IP address
echo "Setting IP address on br0"

brctl setageing br0 0
brctl setfd br0 0
brctl sethello br0 0
brctl stp br0 off

ip addr add $eth0_ip_with_mask dev br0
ip addr add $eth0_ip_with_mask dev eth0
ip link set dev eth0 up
ip link set dev br0 up
ip route add default via $eth0_gateway dev br0
ip route add default via $eth0_gateway dev eth0

echo "Setting IP address on br0 done"

tunctl -t tap0 -u root
brctl addif br0 tap0
ifconfig tap0 0.0.0.0 promisc up

