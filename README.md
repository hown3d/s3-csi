# S3 Implementation for Kubernetes CSI

This repository provides a implementation of the Container Storage Interface by using S3. 

It also includes a FUSE filesystem implementation around S3 in [internal/aws/s3/fs](internal/aws/s3/fs).

## Things that need to be done
- [ ] Implement readonly capability for s3 fs
- [ ] Implement all needed csi endpoints
- [ ] Pass CSI Test Suite (already setup in [test/csi_sanity_test.go](test/csi_sanity_test.go))

## Ideas
- Store information like attachments etc. in s3 metadata