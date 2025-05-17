become campwiz-backend
toolforge build start -L https://gitlab.wikimedia.org/nokibsarkar/campwiz-backend.git --ref $1
toolforge webservice buildservice restart --mount all
toolforge jobs load jobs.yaml
toolforge jobs restart campwiz-task-manager
exit
exit