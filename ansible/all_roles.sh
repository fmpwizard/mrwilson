#!/bin/bash
ansible-playbook  init.yml -i hosts/production --tags=common --ask-vault-pass -u root
ansible-playbook -i hosts/production playbooks/deploy.yml --ask-vault-pass -u root
ansible-playbook mrwilson.yml -i hosts/production  --tags=common --ask-vault-pass -u root
ansible-playbook mrwilson.yml -i hosts/production  --tags=common-service --ask-vault-pass -u root
ansible-playbook mrwilson.yml -i hosts/production  --tags=mrwilson --ask-vault-pass -u root
ansible-playbook -i hosts/production playbooks/site-state.yml --ask-vault-pass --extra-vars "app_state=restart" -u root

## Full
## You can run this on a fresh server or one already running, 100% safe
## We start with init by installing python using raw all the way to copying needed support unit files
## and finally copying the binary and restarting it


# ansible-playbook site.yml -i hosts/production  --tags=full --ask-vault-pass -u root
