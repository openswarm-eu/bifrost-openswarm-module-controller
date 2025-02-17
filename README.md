# CMD
The following command line arguments are used to configure the behaviour
+ -mqtt: the MQTT broker URL
+ -id: ID of the node (has to be unique)
+ energyCommunityId: the ID of the energy community this node is part of

Futhermore, the following flags are used:
+ -l: be part of the leader election cluster
+ -b: boostrap the leader election cluster (use this only on the first node in the leader election cluster)

# Build
```sh
docker build -t pv -f Dockerfile-pv .
docker build -t charger -f Dockerfile-charger .
```

# Run
Start MQTT broker
```sh
docker run -d -p 1883:1883 emqx/emqx
```

Start ONE node as boostrap node for the leader election cluster (e.g PV):
```sh
docker run -it pv -url tcp://host.docker.internal:1883 -l -b
```

The other nodes can optionally be started as part of the leader election cluster:
```sh
docker run -it charger -url tcp://host.docker.internal:1883 -l
docker run -it pv -url tcp://host.docker.internal:1883 -l
```

# MQTT
A PV nodes awaits on the topic `<id>/production` for the following JSON message:
```json
{ "production": 12345678 }
```

A charger nodes sends the follwing charging set point JSON message to topic `<id>/chargingSetPoint`:
```json
{ "chargingSetPoint": 12345678 }
```