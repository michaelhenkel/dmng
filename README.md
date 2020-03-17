# Get the code
```
git clone https://github.com/michaelhenkel/dmng
```

# Build the binaries
```
cd dmng
make
```

# Run a device agent
```
cd dmng/build
./dm_server -port 10000 -name device1
```

# Create some interfaces
```
cd dmng/build
./dm_client -server_addr localhost:10000 -addport eth1,eth4,eth3,eth5,eth7
```

# Check db
```
cat dmng/build/db.json | jq .
```
