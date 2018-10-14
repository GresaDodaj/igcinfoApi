# IGC api

This IGC api is an online service that will allow users to browse information about IGC files. 
IGC is an international file format for soaring track files that are used by paragliders and gliders.The program will store submitted tracks in memory and then they can be browsed.

For the development of the IGC processing is used the open source IGC library for Go: <a href="https://github.com/marni/goigc">goigc</a>



## Getting Started



The root of the Igc Api is /igcinfo/, and if the server is asked about the root it will respond with 404 error.
The rest of the verbs are subsequently attached to the /igcinfo/api/* root


### GET/api    /igcinfo/api

This will return metainformation about the API
The response type is application/json.
Response code:200 if everything is OK.
Body template:

    {
      "uptime": <uptime>
      "info": "Service for IGC tracks."
      "version": "v1"
    }

where: <uptime> is the current uptime of the service formatted according to Duration format as specified by ISO 8601. 


### POST/api/igc    /igcinfo/api/igc

Track registration

Request body template:

    {
      "url": "<url>"
    }


Response will be the id of the track:

    {
      "id": "<id>"
    }



### GET/api/igc   /igcinfo/api/igc

Returns the array of IDs of all the tracked stored in memory, or an empty array if no tracks have been stored yet.
    
    [<id1>, <id2>, ...]
        


### GET/api/igc/<id>      /igcinfo/api/igc/\<id\>

Returns the meta information about a given track with the provided <id>, or NOT FOUND response.
Response code:200 if everything is OK, or 404 if meta information about the requested id is not found.
    {
        "H_date": <date from File Header, H-record>,
        "pilot": <pilot>,
        "glider": <glider>,
        "glider_id": <glider_id>,
        "track_length": <calculated total track length>
    }


### GET/api/igc/<id>/<field>     /igcinfo/api/igc/\<id\>/\<field\>

Returns single meta information about a given track with the provided <id>.
 

Response: 

    <pilot> for pilot
    
    <glider> for glider
    
    <glider_id> for glider_id
    
    <calculated total track length> for track_length
    
    <H_date> for H_date
    



## Deployment

Deployed in Heroku: https://igcapi.herokuapp.com/



