#!/bin/bash

# 三角机构TRPG单人引擎 - Kubernetes部署脚本
# 使用方法: ./deploy-k8s.sh [apply|delete|status]

set -e

ACTION=${1:-apply}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
K8S_DIR="$(cd "$SCRIPT_DIR/../kubernetes" && pwd)"

echo "=========================================="
echo "TRPG Solo Engine - Kubernetes Deployment"
echo "Action: $ACTION"
echo "=========================================="

# 检查kubectl
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is not installed"
    exit 1
fi

# 检查集群连接
if ! kubectl cluster-info &> /dev/null; then
    echo "Error: Cannot connect to Kubernetes cluster"
    exit 1
fi

cd "$K8S_DIR"

case $ACTION in
    apply)
        echo "Deploying to Kubernetes..."
        
        # 创建命名空间
        echo "Creating namespace..."
        kubectl apply -f namespace.yaml
        
        # 创建ConfigMap和Secret
        echo "Creating ConfigMaps and Secrets..."
        kubectl apply -f configmap.yaml
        
        # 提示用户更新secrets
        echo ""
        echo "⚠️  WARNING: Please update secrets.yaml with production values before proceeding!"
        read -p "Have you updated the secrets? (y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "Deployment cancelled. Please update secrets.yaml and try again."
            exit 1
        fi
        
        kubectl apply -f secrets.yaml
        
        # 部署数据库
        echo "Deploying PostgreSQL..."
        kubectl apply -f postgres-deployment.yaml
        
        # 部署Redis
        echo "Deploying Redis..."
        kubectl apply -f redis-deployment.yaml
        
        # 等待数据库就绪
        echo "Waiting for database to be ready..."
        kubectl wait --for=condition=ready pod -l app=postgres -n trpg-solo-engine --timeout=300s
        
        echo "Waiting for Redis to be ready..."
        kubectl wait --for=condition=ready pod -l app=redis -n trpg-solo-engine --timeout=300s
        
        # 部署后端
        echo "Deploying backend..."
        kubectl apply -f backend-deployment.yaml
        
        # 等待后端就绪
        echo "Waiting for backend to be ready..."
        kubectl wait --for=condition=ready pod -l app=trpg-backend -n trpg-solo-engine --timeout=300s
        
        # 部署Ingress
        echo "Deploying Ingress..."
        kubectl apply -f ingress.yaml
        
        # 部署HPA
        echo "Deploying HPA..."
        kubectl apply -f hpa.yaml
        
        echo ""
        echo "✓ Deployment completed successfully!"
        ;;
        
    delete)
        echo "Deleting Kubernetes resources..."
        read -p "Are you sure you want to delete all resources? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            kubectl delete -f hpa.yaml --ignore-not-found=true
            kubectl delete -f ingress.yaml --ignore-not-found=true
            kubectl delete -f backend-deployment.yaml --ignore-not-found=true
            kubectl delete -f redis-deployment.yaml --ignore-not-found=true
            kubectl delete -f postgres-deployment.yaml --ignore-not-found=true
            kubectl delete -f secrets.yaml --ignore-not-found=true
            kubectl delete -f configmap.yaml --ignore-not-found=true
            kubectl delete -f namespace.yaml --ignore-not-found=true
            echo "✓ All resources deleted"
        else
            echo "Deletion cancelled"
        fi
        ;;
        
    status)
        echo "Checking deployment status..."
        echo ""
        echo "Namespace:"
        kubectl get namespace trpg-solo-engine
        echo ""
        echo "Pods:"
        kubectl get pods -n trpg-solo-engine
        echo ""
        echo "Services:"
        kubectl get services -n trpg-solo-engine
        echo ""
        echo "Deployments:"
        kubectl get deployments -n trpg-solo-engine
        echo ""
        echo "Ingress:"
        kubectl get ingress -n trpg-solo-engine
        echo ""
        echo "HPA:"
        kubectl get hpa -n trpg-solo-engine
        echo ""
        echo "PVC:"
        kubectl get pvc -n trpg-solo-engine
        ;;
        
    *)
        echo "Usage: $0 [apply|delete|status]"
        exit 1
        ;;
esac

echo ""
echo "Useful commands:"
echo "  View logs: kubectl logs -f deployment/trpg-backend -n trpg-solo-engine"
echo "  Get pods: kubectl get pods -n trpg-solo-engine"
echo "  Describe pod: kubectl describe pod <pod-name> -n trpg-solo-engine"
echo "  Execute command: kubectl exec -it <pod-name> -n trpg-solo-engine -- /bin/sh"
echo "  Port forward: kubectl port-forward svc/trpg-backend-service 8080:8080 -n trpg-solo-engine"
echo ""
