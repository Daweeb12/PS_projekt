# PS_projekt

## Navodila za minimalno implementacijo

Terminal A (server):
> ./PSMain -p=9000 -id=0

Terminal B (server):
> ./PSMain -p=9001 -id=1

Terminal B.1 (server):
> ./PSMain -p=9002 -id=2
...

Terminal C (client):
> go run ./cmd/client_run -- -url=localhost:9001

Terminal C.1 (client):
> go run ./cmd/client_run -- -url=localhost:9001
...

## Primer funkcije
> createuser <ime>
(Boste dobili userId)

> setuser <userId>

> createtopic <topic_ime>
(Boste dobili topicId)

> postmessage <topicId> <message_content>

> listtopics
(Boste dobili topic list + topicId za vsako)

