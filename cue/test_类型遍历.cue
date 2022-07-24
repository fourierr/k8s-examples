package cue

#a: {
    "hello": "Barcelona"
    "nihao": "Shanghai"
}

for k, v in #a {
    "\(k)": {
        nameLen: len(v)
        value:   v
    }
}