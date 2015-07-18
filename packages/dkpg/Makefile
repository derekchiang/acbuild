.PHONY: all

BUILDDIR := dpkg-build

all:
	rm -rf $(BUILDDIR)
	mkdir -p $(BUILDDIR)/rootfs
	cp manifest $(BUILDDIR)
	while read p; do\
	  mkdir -p $(BUILDDIR)/rootfs$$(dirname $$p); \
	  install -t $(BUILDDIR)/rootfs$$(dirname $$p) -m $$(stat -c "%a" $$p) $$p; \
	done <files.txt
# https://packages.debian.org/sid/amd64/dpkg/filelist
	actool build --overwrite $(BUILDDIR) dpkg.aci
