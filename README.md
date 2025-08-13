# fetch-duck
fetch - duck is a mail scrapper that seeks accounting documents and uploads them to accounting api of (hopefully some day) your choice, but for now, my choice. 

## The Why
As a small business owner, I've been dealing with admin myself with different degrees of success for some time now. And sure I've heard of people that are amazing and maticulous with gathering every piece of receipt from buying a car to paying a cab driver, but I'm not like that, keeping track of menial receipts never was a priority, and yet. it is important. And so fetch-duck comes into play. What's better thah full automation of your gmail invoice stream into one background nice process, that also takes care of uploading it to you neighbor friendly acounting API? Yes, exactly, nothing. And this is how fetch - duck idea was born. 
That, or my wife told me about some app that does, and I retorted, getting this? No way. I'm building it myself. 


### authentication 
For authentication, for my server's API, that is the flow after OAuth 3 legged process, I had a dilemma, to go the JWT root or the stateful session root I thought about it like so, this is a small web server, not a large destributed system, stateful sessions will provide superior security and control, like instantly revoking user sessions, deleting them from the database. And given the fact that the application is manages Gmail API (meaningful data), I prefered server-side security.