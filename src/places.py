from geotext import GeoText
import operator
files = ['../16052538.json',
'../16282819.json',
'../16492994.json',
'../16735011.json',
'../16967543.json',
'../17205865.json',
'../17442187.json',
'../17663077.json',
'../17902901.json',
'../18113144.json',
'../18354503.json',
'../18589702.json']

locations = {}
for file in files:
    with open(file, 'r') as myfile:
        data=myfile.read().replace('\n', '')
    places = GeoText(data)
    for city in places.cities:
        if city not in locations:
            locations[city] = 1
        else:
            locations[city] += 1
sorted_locations =  sorted(locations.items(), key=operator.itemgetter(1), reverse=True)
print(locations)
print(sorted_locations)