# Resume-Operator - Jeff Davis - resume.jefedavis.dev

A Kubernetes operator built with
[operator-builder](https://github.com/nukleros/operator-builder).

This respository and it's companion [Resume](https://github.com/JefeDavis/Resume) serves two purposes:

* Backup of my resume for rapid deployment and development
* Showcasing some skills while I'm at it:
  - Golang
	- Docker
	- Kubernetes
	- Kubernetes Operators
	- CI/CD
	- GitHub Actions
	- MakeFile
	- Hugo Web Framework
	- HTML5, SCSS
	- Writing Great Documentation

# Why
I hate updating resumes, It's not so much writing the content that I dislike. Instead, it's messing with Word and other tools to adjust margins, dealing with columns, page breaks and just formatting in general. So, why not use a data structure language like `YAML` and let programming deal with all of the formating. Plus, the portability and rapid deployments are a big upshot as well.

Now I can store all my expierence, certifications, skills, and projects in kubernetes manifests and let the code handle the rest. I work with Kubernetes and Kuberentes Operators quite a bit, why not practice what I preach? 

## OK, but is a Kubernetes Operator really necessary?
Absolutely not, but it was a fun exercise and a neat talking point. Additionally with the other tools I've built and worked on  ([yot](https://github.com/vmware-tanzu-labs/yaml-overlay-tool) and [operator-builder](https://github.com/nukleros/operator-builder)) I was able to create a fully functional operator in about 30 minutes.


# K8s Architecture Overview
![](./resume-operator.png)

## Local Development & Testing

To install the custom resource/s for this operator, make sure you have a
kubeconfig set up for a test cluster, then run:

    make install

To run the controller locally against a test cluster:

    make run

You can then test the operator by creating the sample manifest/s:

    kubectl apply -f config/samples

To clean up:

    make uninstall

## Deploy the Controller Manager

First, set the image:

    export IMG=myrepo/myproject:v0.1.0

Now you can build and push the image:

    make docker-build
    make docker-push

Then deploy:

    make deploy

To clean up:

    make undeploy

## Companion CLI

To build the companion CLI:

    make build-cli

The CLI binary will get saved to the bin directory.  You can see the help
message with:

    ./bin/resumectl help
