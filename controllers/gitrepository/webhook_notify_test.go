package gitrepository

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/devops/pkg/api/devops/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestWebhookReconciler_notifyGitRepo(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	gitRepo := &v1alpha1.GitRepository{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "ns",
			Name:      "repo",
		},
	}

	type fields struct {
		Client   client.Client
		log      logr.Logger
		recorder record.EventRecorder
	}
	type args struct {
		ns   string
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
		verify  func(assert.TestingT, client.Client)
	}{{
		name: "normal case",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema, gitRepo.DeepCopy()),
		},
		args: args{
			ns:   "ns",
			name: "repo",
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err, i)
			return false
		},
		verify: func(t assert.TestingT, c client.Client) {
			repo := &v1alpha1.GitRepository{}
			err := c.Get(context.TODO(), types.NamespacedName{
				Namespace: "ns",
				Name:      "repo",
			}, repo)
			assert.Nil(t, err)

			assert.NotEmpty(t, repo.Annotations[v1alpha1.AnnotationKeyWebhookUpdates])
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &WebhookReconciler{
				Client:   tt.fields.Client,
				log:      tt.fields.log,
				recorder: tt.fields.recorder,
			}
			tt.wantErr(t, r.notifyGitRepo(tt.args.ns, tt.args.name), fmt.Sprintf("notifyGitRepo(%v, %v)", tt.args.ns, tt.args.name))
			tt.verify(t, tt.fields.Client)
		})
	}
}

func TestWebhookReconciler_notifyGitRepos(t *testing.T) {
	schema, err := v1alpha1.SchemeBuilder.Register().Build()
	assert.Nil(t, err)

	gitRepo := &v1alpha1.GitRepository{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "ns",
			Name:      "repo",
		},
	}
	gitRepoA := &v1alpha1.GitRepository{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "ns",
			Name:      "repo-a",
		},
	}

	type fields struct {
		Client   client.Client
		log      logr.Logger
		recorder record.EventRecorder
	}
	type args struct {
		ns    string
		repos string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
		verify  func(assert.TestingT, client.Client)
	}{{
		name: "normal case",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema,
				gitRepo.DeepCopy(), gitRepoA.DeepCopy()),
		},
		args: args{
			ns:    "ns",
			repos: "repo,repo-a",
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.Nil(t, err, i)
			return false
		},
		verify: func(t assert.TestingT, c client.Client) {
			repo1 := &v1alpha1.GitRepository{}
			err := c.Get(context.TODO(), types.NamespacedName{
				Namespace: "ns",
				Name:      "repo",
			}, repo1)
			assert.Nil(t, err)
			assert.NotEmpty(t, repo1.Annotations[v1alpha1.AnnotationKeyWebhookUpdates])

			repo2 := &v1alpha1.GitRepository{}
			err = c.Get(context.TODO(), types.NamespacedName{
				Namespace: "ns",
				Name:      "repo-a",
			}, repo2)
			assert.Nil(t, err)
			assert.NotEmpty(t, repo2.Annotations[v1alpha1.AnnotationKeyWebhookUpdates])
		},
	}, {
		name: "has errors",
		fields: fields{
			Client: fake.NewFakeClientWithScheme(schema,
				gitRepo.DeepCopy(), gitRepoA.DeepCopy()),
		},
		args: args{
			ns:    "ns",
			repos: "repo,repo-a,repo-b",
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			assert.NotNil(t, err, i)
			return true
		},
		verify: func(t assert.TestingT, c client.Client) {
			repo1 := &v1alpha1.GitRepository{}
			err := c.Get(context.TODO(), types.NamespacedName{
				Namespace: "ns",
				Name:      "repo",
			}, repo1)
			assert.Nil(t, err)
			assert.NotEmpty(t, repo1.Annotations[v1alpha1.AnnotationKeyWebhookUpdates])

			repo2 := &v1alpha1.GitRepository{}
			err = c.Get(context.TODO(), types.NamespacedName{
				Namespace: "ns",
				Name:      "repo-a",
			}, repo2)
			assert.Nil(t, err)
			assert.NotEmpty(t, repo2.Annotations[v1alpha1.AnnotationKeyWebhookUpdates])
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &WebhookReconciler{
				Client:   tt.fields.Client,
				log:      tt.fields.log,
				recorder: tt.fields.recorder,
			}
			tt.wantErr(t, r.notifyGitRepos(tt.args.ns, tt.args.repos), fmt.Sprintf("notifyGitRepos(%v, %v)", tt.args.ns, tt.args.repos))
			tt.verify(t, tt.fields.Client)
		})
	}
}
