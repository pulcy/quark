# Droplets: Machine creator

## Show all DNS records

```
./droplets dns records --domain pulcy.com
```

## Listing instances of a cluster

```
./droplets instances -p vultr --domain pulcy.com --name reg-c21
./droplets instances -p vultr reg-c21.pulcy.com
```

## Add a new instance to an existing cluster

```
./droplets create instance -p vultr --domain iggi.xyz --name alpha-tg2d
./droplets create instance -p vultr alpha-tg2d.iggi.xyz
```

## Removing an instance from an existing cluster

```
./droplets destroy instance -p vultr --domain iggi.xyz --name alpha-tg2d --prefix ldszw7sj
./droplets destroy instance -p vultr ldszw7sj.alpha-tg2d.iggi.xyz
```
