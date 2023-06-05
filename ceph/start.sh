#docker rm -f $(docker ps -a | grep ceph | awk '{print $1}')
docker rm -f ceph-rgw
docker rm -f ceph-mgr
docker rm -f ceph-osd-1
docker rm -f ceph-osd-2
docker rm -f ceph-osd-3
docker rm -f ceph-mon2
docker rm -f ceph-mon1
docker rm -f ceph-mon0
docker network rm ceph-network
rm -r /home/liusong/cephDir

docker network create --driver bridge --subnet 172.20.0.0/16 ceph-network
#显示网络的详细信息
docker network inspect ceph-network
#搭建mon节点
docker run -itd --name ceph-mon0 --network ceph-network --ip 172.20.0.10 -e CLUSTER=ceph -e WEIGHT=1.0 -e MON_IP=172.20.0.10 -e MON_NAME=ceph-mon0 -e CEPH_PUBLIC_NETWORK=172.20.0.0/16 -v /home/liusong/cephDir/ceph:/etc/ceph -v /home/liusong/cephDir/lib/ceph/:/var/lib/ceph/ -v /home/liusong/cephDir/log/ceph/:/var/log/ceph/ ceph/daemon:latest-nautilus mon
docker run -itd --name ceph-mon1 --network ceph-network --ip 172.20.0.11 -e CLUSTER=ceph -e WEIGHT=1.0 -e MON_IP=172.20.0.11 -e MON_NAME=ceph-mon1 -e CEPH_PUBLIC_NETWORK=172.20.0.0/16 -v /home/liusong/cephDir/ceph:/etc/ceph -v /home/liusong/cephDir/lib/ceph/:/var/lib/ceph/ -v /home/liusong/cephDir/log/ceph/:/var/log/ceph/ ceph/daemon:latest-nautilus mon
docker run -itd --name ceph-mon2 --network ceph-network --ip 172.20.0.12 -e CLUSTER=ceph -e WEIGHT=1.0 -e MON_IP=172.20.0.12 -e MON_NAME=ceph-mon2 -e CEPH_PUBLIC_NETWORK=172.20.0.0/16 -v /home/liusong/cephDir/ceph:/etc/ceph -v /home/liusong/cephDir/lib/ceph/:/var/lib/ceph/ -v /home/liusong/cephDir/log/ceph/:/var/log/ceph/ ceph/daemon:latest-nautilus mon
#创建osd节点
docker exec ceph-mon0 ceph auth get client.bootstrap-osd -o /var/lib/ceph/bootstrap-osd/ceph.keyring
#修改配置文件以兼容etx4硬盘
cp ceph.conf /home/liusong/cephDir/ceph
docker run -itd --privileged=true --name ceph-osd-1 --network ceph-network --ip 172.20.0.13 -e CLUSTER=ceph -e WEIGHT=1.0 -e MON_NAME=ceph-mon0 -e MON_IP=172.20.0.10 -e OSD_TYPE=directory -v /home/liusong/cephDir/ceph:/etc/ceph -v /home/liusong/cephDir/lib/ceph/:/var/lib/ceph/ -v /home/liusong/cephDir/lib/ceph/osd/1:/var/lib/ceph/osd -v /etc/localtime:/etc/localtime:ro ceph/daemon:latest-nautilus osd
docker run -itd --privileged=true --name ceph-osd-2 --network ceph-network --ip 172.20.0.14 -e CLUSTER=ceph -e WEIGHT=1.0 -e MON_NAME=ceph-mon0 -e MON_IP=172.20.0.10 -e OSD_TYPE=directory -v /home/liusong/cephDir/ceph:/etc/ceph -v /home/liusong/cephDir/lib/ceph/:/var/lib/ceph/ -v /home/liusong/cephDir/lib/ceph/osd/2:/var/lib/ceph/osd -v /etc/localtime:/etc/localtime:ro ceph/daemon:latest-nautilus osd
docker run -itd --privileged=true --name ceph-osd-3 --network ceph-network --ip 172.20.0.15 -e CLUSTER=ceph -e WEIGHT=1.0 -e MON_NAME=ceph-mon0 -e MON_IP=172.20.0.10 -e OSD_TYPE=directory -v /home/liusong/cephDir/ceph:/etc/ceph -v /home/liusong/cephDir/lib/ceph/:/var/lib/ceph/ -v /home/liusong/cephDir/lib/ceph/osd/3:/var/lib/ceph/osd -v /etc/localtime:/etc/localtime:ro ceph/daemon:latest-nautilus osd
#搭建mgr节点
docker run -itd --privileged=true --name ceph-mgr --network ceph-network --ip 172.20.0.16 -e CLUSTER=ceph -p 7000:7000 --pid=container:ceph-mon0 -v /home/liusong/cephDir/ceph:/etc/ceph -v /home/liusong/cephDir/lib/ceph/:/var/lib/ceph/ ceph/daemon:latest-nautilus mgr
docker exec ceph-mgr ceph mgr module enable dashboard
#搭建rgw节点
docker exec ceph-mon0 ceph auth get client.bootstrap-rgw -o /var/lib/ceph/bootstrap-rgw/ceph.keyring
docker run -itd --privileged=true --name ceph-rgw --network ceph-network --ip 172.20.0.17 -e CLUSTER=ceph -e RGW_NAME=ceph-rgw -p 7480:7480 -v /home/liusong/cephDir/lib/ceph/:/var/lib/ceph/ -v /home/liusong/cephDir/ceph:/etc/ceph -v /etc/localtime:/etc/localtime:ro ceph/daemon:latest-nautilus rgw
#检查Ceph状态
docker exec ceph-mon0 ceph -s
#添加rgw用户
docker exec -it ceph-rgw radosgw-admin user create --uid="test" --display-name="test user"