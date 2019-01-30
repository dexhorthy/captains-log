#docker_build("captains-log", ".")
repo = local_git_repo('.')

(fast_build('dexhorthy/captains-log', 'Dockerfile-tilt', '/root/manager')
 .add(repo.path('/'), '/go/src/github.com/dexhorthy/captains-log/')
 .run('CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install github.com/dexhorthy/captains-log/cmd/manager')
 .run('mv /go/bin/manager /manager'))

k8s_yaml(kustomize("./config/default"))
