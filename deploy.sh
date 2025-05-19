become campwiz-backend
toolforge build start -L --image-name campwiz-backend-fake-nginx https://gitlab.wikimedia.org/nokibsarkar/campwiz-backend.git --ref fake-nginx
toolforge build start -L https://gitlab.wikimedia.org/nokibsarkar/campwiz-backend.git 
toolforge webservice buildservice restart --mount all --buildservice-image tool-campwiz-backend/campwiz-backend-fake-nginx:latest
toolforge jobs load jobs.yaml
toolforge jobs restart campwiz-task-manager
toolforge jobs restart campwiz-backend-readonly
toolforge jobs restart campwiz-backend-thing
exit
exit