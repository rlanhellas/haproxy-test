#!/bin/bash
servera="servera"
serverb="serverb"
servera_pri=0
serverb_pri=1
version=`date +%s`
current=$(cat ./current)

if [ $current == $servera ]; then
    servera_pri=1
    serverb_pri=0
    echo $serverb > ./current
elif [ $current == $serverb ]; then
    servera_pri=0
    serverb_pri=1
    echo $servera > ./current
fi


sed -e "s/#PRIA#/$servera_pri/g; \
        s/#PRIB#/$serverb_pri/g; \
        s/#VERSION#/$version/g" db.internal.template > db.internal