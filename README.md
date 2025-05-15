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

