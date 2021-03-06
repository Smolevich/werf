{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}
Build then push images from werf.yaml.

Images will be tagged automatically with the names REPO/IMAGE_NAME:TAG. These tags will be deleted 
after push. See more info about images naming: 
https://flant.github.io/werf/reference/registry/image_naming.html.

The result of bp command is a stages cache for images and named images pushed into the docker 
registry.

If one or more IMAGE_NAME parameters specified, werf will build and push only these images from 
werf.yaml.

{{ header }} Syntax

```bash
werf bp [IMAGE_NAME...] [options]
```

{{ header }} Options

```bash
      --dir='':
            Change to the specified directory to find werf.yaml config
  -h, --help=false:
            help for bp
      --home-dir='':
            Use specified dir to store werf cache files and dirs (use ~/.werf by default)
      --introspect-before-error=false:
            Introspect failed stage in the clean state, before running all assembly instructions 
            of the stage
      --introspect-error=false:
            Introspect failed stage in the state, right after running failed assembly instruction
      --pull-password='':
            Docker registry password to authorize pull of base images
      --pull-username='':
            Docker registry username to authorize pull of base images
      --push-password='':
            Docker registry password to authorize push to the docker repo
      --push-username='':
            Docker registry username to authorize push to the docker repo
      --registry-password='':
            Docker registry password to authorize pull of base images and push to the docker repo
      --registry-username='':
            Docker registry username to authorize pull of base images and push to the docker repo
      --repo='':
            Docker repository name to push images to. CI_REGISTRY_IMAGE will be used by default if 
            available.
      --ssh-key=[]:
            Enable only specified ssh keys (use system ssh-agent by default)
      --tag=[]:
            Add tag (can be used one or more times)
      --tag-branch=false:
            Tag by git branch
      --tag-build-id=false:
            Tag by CI build id
      --tag-ci=false:
            Tag by CI branch and tag
      --tag-commit=false:
            Tag by git commit
      --tmp-dir='':
            Use specified dir to store tmp files and dirs (use system tmp dir by default)
```

{{ header }} Environments

```bash
  $WERF_ANSIBLE_ARGS                
  $WERF_DOCKER_CONFIG               
  $WERF_IGNORE_CI_DOCKER_AUTOLOGIN  
  $WERF_HOME                        
  $WERF_TMP                         
```

