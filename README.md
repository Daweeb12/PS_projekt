# PS_projekt

## Navodila za minimalno implementacijo
./compile.sh
go build . 
Terminal A (server):
> ./PS_projekt -p=9000 -id=0

Terminal B (server):
> ./PS_projekt -p=9001 -id=1

Terminal B.1 (server):
> ./PS_projekt -p=9002 -id=n  , n < 5 


Terminal C (client):
./PS_projekt -id=n , n >= 5 
Terminal C.1 (client):

## Primer funkcije
> createuser <ime>
(Boste dobili userId)

> setuser <userId>

> createtopic <topic_ime>
(Boste dobili topicId)

> postmessage <topicId> <message_content>

> listtopics
(Boste dobili topic list + topicId za vsako)

