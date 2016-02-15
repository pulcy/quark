# Quark: Cluster & instance creator

## Show all DNS records

```
quark dns records --domain pulcy.com
```

## Listing instances of a cluster

```
quark instance list -p vultr --domain pulcy.com --name c47
# or
quark instance list -p vultr c47.pulcy.com
```

## Creating a new cluster in vagrant

```
quark cluster create -p vagrant --domain pulcy.com
```

## Add a new instance to an existing cluster

```
quark instance create -p vultr --domain iggi.xyz --name a75
# or
quark instance create -p vultr a75.iggi.xyz
```

## Removing an instance from an existing cluster

```
quark instance destroy -p vultr --domain iggi.xyz --name a75 --prefix ldszw7sj
# or
quark instance destroy -p vultr ldszw7sj.a75.iggi.xyz
```
