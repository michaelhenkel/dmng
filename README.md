# Get the code
```
git clone https://github.com/michaelhenkel/dmng
```

# Build the binaries
```
cd dmng
make
```

# Run the FM server
```
cd dmng/build
./fm_server
```

# in a different terminal connect first device
```
cd dmng/build
./dm_server -name device1
```

# in a different terminal connect second device
```
cd dmng/build
./dm_server -name device2
```

# in a different terminal create a fabric
```
cd dmng/build
./fm_client -request request.yaml
```
