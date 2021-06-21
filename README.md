##Mr. Wilson

Personal project, nothing to see yet.


## Setup

You will need to create encrypted files with the staging or production secret values.

1. Duplicate the file `ansible/vars/production.yml` and name it `ansible/vars/production_vault.yml`
2. Edit the new file so the keys become the values of the <mode>.yml file:  `vault_wit_access_token` : "real wit access token here".
3. Repeat for all variables.
4. Inside the `vars/` folder, run `ansible-vault encrypt production_vault.yml` to encrypt the vault file
5. Enter a new password to encrypt this file, and make sure to keep this password safe in your password manager. You'll need it for deployments.
6. Now you can go into the all_roles.sh file and run the different playbooks to deploy this service.