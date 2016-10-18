# Quark: Pulcy cluster & instance creator

Quark is used to create new clusters (of machines), add machines to existing clusters or
remove machines from existing clusters.

The generated clusters are configured such that fleet jobs can be scheduled on them as soon
as Quark has finished.

## Building 

```
go get -u github.com/pulcy/quark
```

or 

```
git clone https://github.com/pulcy/quark.git
cd quark
make 
```

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
