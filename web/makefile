all: assetsAll 3rdpartyAll

STATICREADME="GENERATED BY MAKEFILE DO NOT MODIFY DIRECTLY.\
Assets contained here are either copied from the same named directories from the parent folder\
or generated from another source as dictated by the makefile\
to make modifications to any files please see the equivilant folders in the parent dir and the makefile\
\
if you are a curious user of the site and you found this file from messing around then hello i guess"

static:
	mkdir static
	echo $(STATICREADME) > static/README_DO_NOT_MODIFY

assetsAll: assets

assets: static
	cp -r assets static/assets

3rdpartyAll: 3rdparty purecss

3rdparty: static
	cp -r 3rdparty static/3rdparty

purecss: static
	cp -r ../node_modules/purecss/build static/3rdparty/purecss/

clean:
	rm -r static
