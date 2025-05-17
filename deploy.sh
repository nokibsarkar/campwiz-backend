become campwiz-backend
toolforge build start -L https://gitlab.wikimedia.org/nokibsarkar/campwiz-backend.git --ref $1
toolforge webservice buildservice restart --mount all
exit
exit