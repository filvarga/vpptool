vpp=vpp-run
docker cp ./conf/startup.conf $vpp:/etc/startup.conf
docker cp ./conf/vpp.conf $vpp:/etc/vpp.conf
docker exec -e START_VPP=1 -it $vpp /scripts/start
