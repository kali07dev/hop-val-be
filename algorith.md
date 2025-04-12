### approach to solution

1 - fetch data
2- manually check the structure
3- based off struct, create GORM models to match
3.1 - get db configurations from config.yaml filke

Auto-Processes 

4- fetch data:
5- to prevent duplicate entries
    - use primary key to verify entry
    - DB rules will prevent this

    other approach
        - verify key does not exist by fetching using primary key
        - if no record is found
            - proceed with insert

        - if record found
            - return useful error that says this record exists

NB: if Our Data is in array format
    - use the same process
    - but use a loop and for each error, don't kill the process
    - rather store the errors in an array 
    - then return them all

    - this will help in not nulfying all the records if let's say 1 record already exists, that one only won't be processed.

6 - on Read - have read all and by ID (one)
    6.1 - have pagination on read all
    6.2 - have filters based off the data
7 - have search endpoint

if time is available:
    - use cobra to set up the migration and server commands