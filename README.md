liftCA
======

liftCA is a fast, pragmatic and minimalistic web-based TLS certificate management tool.
 
It was built to simplify the work of generating certificates in an enclosed lab environment where a TLS certificate authority setup is required, but where learning to operate complex TLS tools is not the objective.

Please be sure to understand the security implications of trusting new certificate authorities or certificates within the environment you use liftCA in.  Playing with these settings without fully understanding the implications can be dangerous.

Demo Site
---------

Visit http://liftca.com/ to try it out.

How to Install and Run
----------------------

liftCA is available in source form at https://github.com/jeanfric/liftca.  

To build this tool, install Go (http://golang.org/), clone the repository, place it in your `GOPATH`, then run `go build` in `src/liftca/cmd/liftca`.  The resulting `liftca` binary can then be started from that directory.

```
$ git clone https://github.com/jeanfric/liftca
$ export GOPATH="$(pwd)/liftca:$GOPATH"
$ cd liftca/src/liftca/cmd/liftca
$ go get
$ go build
$ ./liftca
2014/07/06 13:37:00 liftCA engaged at ':8080', data file 'store.gob'
```  

License
-------

Copyright 2014 Jean-Francois Richard

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
