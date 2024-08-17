# gomyvip

A silly little script to try to find/print all the current rewards on myvip.co. Outputs the data in a json blob sorted by price.

```
C:\Users\csm10495\Desktop\gomyvip\gomyvip>go run .
[
    {
        "Name": "Free Beer Upgrade",
        "Price": 1000,
        "Description": "Enjoy a free 42oz beer upgrade at TAP Sports bar.",
        "Stock": 34,
        "Partner": "MGM National Harbor"
    },
    {
        "Name": "Complimentary Guidebook",
        "Price": 1000,
        "Description": "Enjoy a complimentary guidebook with the purchase of a regular priced ticket at \nSEA LIFE Arizona Aquarium!\n\nExplore 26,000 square feet of marine wonders at SEA LIFE Arizona. Discover thousands of aquatic creatures, interactive touch pools, and a 360° ocean tunnel!",
        "Stock": 23,
        "Partner": "SEA LIFE Arizona"
    },
...
]
```

If the Stock field is -1, there was no available stock data from the API.
