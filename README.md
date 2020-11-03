# lxcer
Backup/restore tool for `LXC` containers to `restic` with `zstd` compression.

### Requirements

- zstd >= 1.38 
- lxc >= 3.0.3 
- restic >= 0.10.0 

### Description
This is wrapper for `lxc`, `zstd`, `restic` CLI interfaces. Before running: 
- all listed above programs are installed and in the $PATH
- configure the `conf.yml` accordingly
- all hosts that are listed in `conf.yml` are indeed available for lxc cli: `lxc remote list`
- all restic repos that are listed in `conf.yml` are created and not broken

It does simple `restic check` before any work starts and fatals out in case an error. 

#### Backup 
Follows logic below: 
1. Create (remote) snapshot 
2. Publish (remote) snapshot as image
3. Export image as .tar
4. Compress .tar to .tar.zst
5. Push .tar.zst to listed restic repos. 

If run concurrently, then for each remote host starts its own goroutine which creates and publishes snapshots. Then passes image to next goroutine which exports it, then passes to next one which compresses it and passes it further to goroutines that push compressed archives to restic repos. 

#### Restore
Follows logic below: 
1. Download latest snapshot for container
2. Decompress it from .tar.zst to .tar
3. Import .tar as (remote) image
4. Start (remote) container from (remote) image.
