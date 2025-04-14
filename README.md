# campwiz-backend
[![codecov](https://codecov.io/github/nokibsarkar/campwiz-backend/graph/badge.svg?token=E3NPCJRDG3)](https://codecov.io/github/nokibsarkar/campwiz-backend)

# Milestone 15-04-2025: Reduce Import Time from 1 hour to 1 minute
I used GRPC to make the import more distributed. Also, switched from API to Wikimedia Commons replica database with tunneling via developer account. Also, the descriptions are now fetching in background and not blocking the import. The import time is now reduced to 1 minute.
