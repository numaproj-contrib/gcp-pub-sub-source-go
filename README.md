# GCP Pub/Sub Source for Numaflow


The GCP Pub/Sub Source is a custom user-defined source for [Numaflow](https://numaflow.numaproj.io/) that enables the integration of Google Cloud Pub/Sub as a source within your Numaflow pipelines.

- [Quick Start](#quick-start)
- [Using GCP Pub/Sub Source in Your Numaflow Pipeline](#using-gcp-pubsub-source-in-your-numaflow-pipeline)
- [Environment Variables Configuration](#using-environment-variables-to-specify-the-gcp-pubsub-source-configuration)
- [Debugging GCP Pub/Sub Source](#debugging-gcp-pubsub-source)

## Quick Start
This quick start guide will walk you through setting up a GCP Pub/Sub source in a Numaflow pipeline.

### Prerequisites
* [Install Numaflow on your Kubernetes cluster](https://numaflow.numaproj.io/quick-start/)
* Access to a Google Cloud Platform project with Pub/Sub enabled

### Step-by-Step Guide

#### 1. Create a GCP Pub/Sub Topic and Subscription

Using the Google Cloud Console or the `gcloud` command-line tool, create a new Pub/Sub topic and subscription.

#### 2. Deploy a Numaflow Pipeline with GCP Pub/Sub Source

Save the following Kubernetes manifest to a file (e.g., `pubsub-source-pipeline.yaml`), modifying the project ID, topic ID, subscription ID, and the Pub/Sub emulator host accordingly:

```yaml
apiVersion: numaflow.numaproj.io/v1alpha1
kind: Pipeline
metadata:
  name: pubsub-source-pipeline
spec:
  vertices:
    - name: pubsub-source
      source:
        udsource:
          container:
            image: "quay.io/numaio/numaflow-go/gcp-pubsub-source-go:latest"
            env:
              - name: PROJECT_ID
                value: "your-gcp-project-id"
              - name: TOPIC_ID
                value: "your-pubsub-topic-id"
              - name: SUBSCRIPTION_ID
                value: "your-pubsub-subscription-id"
              - name: PUBSUB_EMULATOR_HOST
                value: "gcloud-pubsub.numaflow-system.svc.cluster.local:8681"
    - name: log-sink
      sink:
        log: {}
  edges:
    - from: pubsub-source
      to: log-sink
```

Then apply it to your cluster:
```bash
kubectl apply -f pubsub-source-pipeline.yaml
```

#### 3. Verify the Pipeline

Check if the pipeline is running:
```bash
kubectl get pipeline pubsub-source-pipeline
```

#### 4. Publish Messages to the GCP Pub/Sub Topic

Use the Google Cloud Console or `gcloud` CLI to publish messages to your Pub/Sub topic.

#### 5. Verify the Log Sink

Check the logs of the `log-sink` vertex to see if it received the message from Pub/Sub:
```bash
kubectl logs <log-sink-pod-name>
```

You should see output similar to:
```
Message received from Pub/Sub
```

#### 6. Clean Up

To delete the Numaflow pipeline:
```bash
kubectl delete pipeline pubsub-source-pipeline
```

To delete the Pub/Sub topic and subscription:
```bash
# Replace <topic-id> and <subscription-id> with your topic and subscription IDs
gcloud pubsub topics delete <topic-id>
gcloud pubsub subscriptions delete <subscription-id>
```

Congratulations! You have successfully set up a GCP Pub/Sub source in a Numaflow pipeline.

## Using GCP Pub/Sub Source in Your Numaflow Pipeline

**Prerequisites:**
- Ensure Numaflow is installed on your Kubernetes cluster.
- Access to Google Cloud Platform project with Pub/Sub enabled.

**Step-by-Step Guide:**

1. **Create a GCP Pub/Sub Topic and Subscription:**
   - Use the Google Cloud Console or `gcloud` CLI to create a new Pub/Sub topic and subscription.

2. **Deploy a Numaflow Pipeline with GCP Pub/Sub Source:**
   - Prepare a Kubernetes manifest file, let’s say `pubsub-source-pipeline.yaml`. Here's a template to get you started:

```yaml
apiVersion: numaflow.numaproj.io/v1alpha1
kind: Pipeline
metadata:
  name: pubsub-source-pipeline
spec:
  vertices:
    - name: pubsub-source
      source:
        udsource:
          container:
            image: "quay.io/numaio/numaflow-go/gcp-pubsub-source-go:latest"
            env:
              - name: PROJECT_ID
                value: "your-gcp-project-id"
              - name: TOPIC_ID
                value: "your-pubsub-topic-id"
              - name: SUBSCRIPTION_ID
                value: "your-pubsub-subscription-id"
              - name: PUBSUB_EMULATOR_HOST
                value: "gcloud-pubsub.numaflow-system.svc.cluster.local:8681"
    - name: log-sink
      sink:
        log: {}
  edges:
    - from: pubsub-source
      to: log-sink
    
```



- Modify the environment variables as per your Google Cloud project configuration.
- Apply the configuration to your cluster by running:

  ```bash
  kubectl apply -f pubsub-source-pipeline.yaml
By following these steps, you'll have a GCP Pub/Sub topic and subscription as sources for your Numaflow pipeline running in Kubernetes. Messages from Pub/Sub will be fetched by the pipeline and passed to the log sink for further processing.

## Using Environment Variables to Specify the GCP Pub/Sub Source Configuration

To configure the GCP Pub/Sub source in your Numaflow pipeline using environment variables, you can follow this process:

**Using Environment Variables for GCP Pub/Sub Source Configuration:**

1. **Set Up Environment Variables:**
   - Define environment variables on your Kubernetes deployment to hold the necessary GCP Pub/Sub configuration, such as the project ID, topic ID, subscription ID, and Pub/Sub emulator host.
   - For example, set `PROJECT_ID`, `TOPIC_ID`, `SUBSCRIPTION_ID`, and `PUBSUB_EMULATOR_HOST` as environment variables in your pipeline deployment.

2. **Modify Your Pipeline Configuration:**
   - In your Numaflow pipeline definition (`pubsub-source-pipeline.yaml`), you will reference these environment variables. Ensure your source container in the pipeline specification is configured to use these variables.

Here's an example snippet for your Kubernetes manifest file:

```yaml
apiVersion: numaflow.numaproj.io/v1alpha1
kind: Pipeline
metadata:
  name: pubsub-source-pipeline
spec:
  vertices:
    - name: pubsub-source
      source:
        udsource:
          container:
            image: "quay.io/numaio/numaflow-go/gcp-pubsub-source-go:latest"
            env:
              - name: PROJECT_ID
                value: "your-gcp-project-id"
              - name: TOPIC_ID
                value: "your-pubsub-topic-id"
              - name: SUBSCRIPTION_ID
                value: "your-pubsub-subscription-id"
              - name: PUBSUB_EMULATOR_HOST
                value: "gcloud-pubsub.numaflow-system.svc.cluster.local:8681"
```

3. **Deploy the Configuration:**
   - Apply the updated pipeline configuration to your Kubernetes cluster using the command:

     ```bash
     kubectl apply -f pubsub-source-pipeline.yaml
     ```

By configuring your pipeline in this manner, it allows the GCP Pub/Sub source to dynamically retrieve the configuration from the environment variables, making it easier to manage and change settings.

## Debugging GCP Pub/Sub Source

To enable debugging, set the `DEBUG` environment variable to `true` within the GCP Pub/Sub source container. This will output additional log messages for troubleshooting. See [Debugging](https://numaflow.numaproj.io/development/debugging/) for more information.

## Additional Resources

For more information on Numaflow and how to use it to process data in a Kubernetes-native way, visit the [Numaflow Documentation](https://numaflow.numaproj.io/). For GCP Pub/Sub specific configuration, refer to the [Google Cloud Pub/Sub Documentation](https://cloud.google.com/pubsub/docs).
```

