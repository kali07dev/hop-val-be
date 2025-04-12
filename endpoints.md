Some Endpoints for easy testing by copying and pasting them in insomnia or postman or your api tester

Testing Filters (/api/v1/properties)
/api/v1/properties?ownerName=Moyenda%20Amosi
/api/v1/properties?constructionStage=completed
/api/v1/properties?listingType=Rent
/api/v1/properties?minPrice=290000
/api/v1/properties?maxPrice=310000
/api/v1/properties?minPrice=250000&maxPrice=350000
/api/v1/properties?district=BLANTYRE%20URBAN
/api/v1/properties?area=Zingwangwa
/api/v1/properties?agentId=22
/api/v1/properties?listingType=Rent&district=BLANTYRE%20URBAN
/api/v1/properties?ownerName=Moyenda%20Amosi&constructionStage=completed&listingType=Rent&minPrice=290000&maxPrice=310000&district=BLANTYRE%20URBAN&area=Zingwangwa&agentId=22
/api/v1/properties?listingType=Sale (Example of filter NOT matching)
/api/v1/properties?minPrice=500000 (Example of filter NOT matching)
/api/v1/properties?district=Lilongwe (Example of filter NOT matching)

Testing Search (/api/v1/properties/search)
/api/v1/properties/search?q=Moyenda
/api/v1/properties/search?q=Amosi
/api/v1/properties/search?q=Detached
/api/v1/properties/search?q=completed
/api/v1/properties/search?q=dream%20home
/api/v1/properties/search?q=Zingwangwa
/api/v1/properties/search?q=BLANTYRE
/api/v1/properties/search?q=Mondoni
/api/v1/properties/search?q=Aubrey
/api/v1/properties/search?q=Namfuko
/api/v1/properties/search?q=Residential
/api/v1/properties/search?q=tank
/api/v1/properties/search?q=parking
/api/v1/properties/search?q=xyzNonExistent123 (Example of search NOT matching)