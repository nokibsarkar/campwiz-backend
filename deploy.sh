become campwiz-backend
toolforge build start -L --image-name campwiz-backend-fake-nginx https://gitlab.wikimedia.org/nokibsarkar/campwiz-backend.git --ref fake-nginx
toolforge webservice buildservice restart --mount all --buildservice-image tool-campwiz-backend/campwiz-backend-fake-nginx:latest
exit
exit