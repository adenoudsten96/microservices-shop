import requests
import json

url = "http://localhost:30002/product"

products = [{
    "SKU": "SKU1",
	"Name": "Raspberry Pi",
	"Price": 3000,
	"Description": "A small computer."
},{
    "SKU": "SKU2",
	"Name": "Arduino",
	"Price": 1500,
	"Description": "An even smaller computer."
},
{
    "SKU": "SKU3",
	"Name": "Resistor",
	"Price": 100,
	"Description": "Resists stuff."
},
{
    "SKU": "SKU4",
	"Name": "Mouse",
	"Price": 2000,
	"Description": "Meep."
},
{
    "SKU": "SKU5",
	"Name": "Keyboard",
	"Price": 6000,
	"Description": "For typing."
},
{
    "SKU": "SKU6",
	"Name": "Monitor",
	"Price": 10000,
	"Description": "For your eyeballs."
}
]

for p in products:
    a = requests.post(url, json=p)
    print(a.status_code)
    print(a.content)