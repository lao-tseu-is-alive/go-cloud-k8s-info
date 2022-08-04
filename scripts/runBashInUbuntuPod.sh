#!/bin/bash
echo "about to enter an interactive session that will end after you exit the bash terminal"
kubectl -n default run -i --tty bash --image=ubuntu --restart=Never -- bash