#!/bin/bash


if [ "$1" = "m" ] ; then 
    ./PS_projekt 
elif [ "$1" = "c" ];then
    ./PS_projekt -id=10
fi
