gx:
	go get github.com/whyrusleeping/gx
	go get github.com/whyrusleeping/gx-go

<<<<<<< HEAD
deps: gx
=======
covertools:
	go get github.com/mattn/goveralls
	go get golang.org/x/tools/cmd/cover

deps: gx covertools
>>>>>>> 258d5c409a01370dfe542ceadc3d1669659150fe
	gx --verbose install --global
	gx-go rewrite

publish:
	gx-go rewrite --undo

