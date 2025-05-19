# campwiz-backend
[![codecov](https://codecov.io/github/nokibsarkar/campwiz-backend/graph/badge.svg?token=E3NPCJRDG3)](https://codecov.io/github/nokibsarkar/campwiz-backend)

# Milestones
## 15-04-2025: Reduce Import Time from 1 hour to 1 minute
I used GRPC to make the import more distributed. Also, switched from API to Wikimedia Commons replica database with tunneling via developer account. Also, the descriptions are now fetching in background and not blocking the import. The import time is now reduced to 1 minute.
## 16-05-2025: Deployed to Toolforge (almost)
After a lot of effort, I managed to
- Create another tool named `campwiz-backend` for backend hosting
- Added build service for golang project using `1.24.1` version (Thanks to @dcaro and @dhinus)
- Access all the configuration file by adding `/data/project/campwiz-backend` into viper path (Thanks to @dcaro)
- Get response by binding to `0.0.0.0` instead of `localhost` (Thanks to @jeremy_b)
My new issue is
- My code is built for nodejs version 22, but toolforge supports upto 18. That's why some functions are not available like `toSorted`

## 19-05-2025: The tools is frequently going to `read-only` mode
Last night, I accidentally remove the flag for  port, so it was running the default port. But since two servers were running on the same default port, one of them errored. Now nginx was trying to balance the load between them. But since one of them errored, nginx was switching to the read-only mode. Now I added the port flag.

## 20-05-2025: The tools is now running behind nginx reverse proxy on toolforge
So, David Caro helped me setup the architecture I wanted. What I wanted was to run the golang server behind nginx reverse proxy.Also, another read-only server would be used as failover server. So, if the main server goes down, the nginx will switch to the read-only server. I also started running a GRPC Server on port 50051. The GRPC server would be used for asynchronous tasks like importing images and descriptions. The GRPC server is not behind the nginx reverse proxy. So, it can be accessed directly. Our solution was (actually, david caror's solution):
- Build a image with standard golang buildpack.
    - It built two binaries: `campwiz` and `taskmanager`
- Build another image with `heroku-php-nginx` buildpack from the branch `fake-nginx`
    - Actually, it just contains a `composer.json` file which tricks the build pack into choosing the `heroku-php-nginx` buildpack.
    - Also contains a `nginx.conf` file which is used to configure the nginx server.
    - This `nginx.conf` file is given in the `web` entry of the `Procfile`.
- Now, when both of the images are built, 
    - we first run three continuous jobs using the golang built image:
        - `campwiz` which runs the main server on port 8080
        - `taskmanager` which runs the GRPC server on port 50051
        - `campwiz -readonly` which runs the read-only server on port 8081 that is used for failover
    - Then we run the webservice using our hackish `heroku-php-nginx` image.
        - Now the nginx is forwarding the requests to the main server on port 8080 and the read-only server on port 8081. The GRPC server is running on port 50051.
But my current issue is that, I see some performance hit after this change. I am not sure if it is because of the nginx reverse proxy or something else.