#!/bin/bash
echo "about to prepare a script to run trivy reports based on list of images in images_list.txt"
for i in `cat images_list.txt`;do echo "trivy image  --severity MEDIUM,HIGH,CRITICAL --format json --output trivy_security_scan_"`basename $i`.json" $i";done >make_scan.sh
for i in `cat images_list.txt`;do echo "trivy image  --severity MEDIUM,HIGH,CRITICAL --format table --output trivy_security_scan_"`basename $i`.txt" $i";done >>make_scan.sh

