# themis fake (only for local development)

This project acts as a fake of themis. If your application depends on themis, instead of connecting to a real themis endpoint, you can deploy this project to a local k3d on your computer.

# Deploy themis fake to k3d

1. Pre-requisite
   
   1. ```
      brew install docker
      brew install k3d
      ```

2. As below:
   
   1. ```
      k3d create \
          --publish 8080:80 \
          --publish 8443:443 \
          --workers 3 \
          --server-arg --no-deploy=traefik
      
      k3d get-kubeconfig --name='k3s-default'
      
      kubectl config use-context default
      
      ```

3. To install **themis-fake**
   
   1. ```
      # import once only
      make k3dimport
      
      kubectl apply -f deploy/themis.yaml
      ```
