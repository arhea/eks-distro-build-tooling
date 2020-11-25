/*
Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License"). You may
not use this file except in compliance with the License. A copy of the
License is located at

    http://aws.amazon.com/apache2.0/

or in the "license" file accompanying this file. This file is distributed
on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
express or implied. See the License for the specific language governing
permissions and limitations under the License.
*/
package pkg

import (
	"fmt"
	"path"

	eksDistrov1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
)

// GetKubernetesComponent returns the Component for Kubernetes
func (r *ReleaseConfig) GetKubernetesComponent(spec eksDistrov1alpha1.ReleaseSpec) (*eksDistrov1alpha1.Component, error) {
	kgv, err := newKubeGitVersionFile(r.BuildRepoSource, spec.Channel)
	if err != nil {
		return nil, err
	}
	gitTag, err := r.readK8sTag(r.BuildRepoSource, spec.Channel)
	if err != nil {
		return nil, err
	}
	assets := []eksDistrov1alpha1.Asset{}

	osComponentMap := map[string][]string{
		"linux":   []string{"client", "server", "node"},
		"windows": []string{"client", "node"},
		"darwin":  []string{"client"},
	}
	osArchMap := map[string][]string{
		"linux":   []string{"arm64", "amd64"},
		"windows": []string{"amd64"},
		"darwin":  []string{"amd64"},
	}
	osBinaryMap := map[string][]string{
		"linux": []string{
			"kube-apiserver",
			"kube-controller-manager",
			"kube-proxy",
			"kube-scheduler",
			"kubectl",
			"kubelet",
		},
		"darwin": []string{
			"kubectl",
		},
		"windows": []string{
			"kube-proxy.exe",
			"kubeadm.exe",
			"kubectl.exe",
			"kubelet.exe",
		},
	}
	binaryAssets := []eksDistrov1alpha1.Asset{}

	for os, arches := range osArchMap {
		for _, arch := range arches {
			for _, binary := range osBinaryMap[os] {
				filename := path.Join("bin", os, arch, binary)
				sha256, sha512, err := r.ReadK8sShaSums(spec.Channel, filename)
				if err != nil {
					return nil, err
				}
				binaryAssets = append(binaryAssets, eksDistrov1alpha1.Asset{
					Name:        filename,
					Type:        "Archive",
					Description: fmt.Sprintf("%s binary for %s/%s", binary, os, arch),
					OS:          os,
					Arch:        []string{arch},
					Archive: &eksDistrov1alpha1.AssetArchive{
						Path: path.Join(
							fmt.Sprintf("kubernetes-%s", spec.Channel),
							"releases",
							fmt.Sprintf("%d", spec.Number),
							"artifacts",
							"kubernetes",
							gitTag,
							filename,
						),
						SHA512: sha512,
						SHA256: sha256,
					},
				})
			}
			for _, component := range osComponentMap[os] {
				filename := fmt.Sprintf("kubernetes-%s-%s-%s.tar.gz", component, os, arch)
				sha256, sha512, err := r.ReadK8sShaSums(spec.Channel, filename)
				if err != nil {
					return nil, err
				}
				assets = append(assets, eksDistrov1alpha1.Asset{
					Name:        filename,
					Type:        "Archive",
					Description: fmt.Sprintf("Kubernetes %s tarball for %s/%s", component, os, arch),
					OS:          os,
					Arch:        []string{arch},
					Archive: &eksDistrov1alpha1.AssetArchive{
						Path: path.Join(
							fmt.Sprintf("kubernetes-%s", spec.Channel),
							"releases",
							fmt.Sprintf("%d", spec.Number),
							"artifacts",
							"kubernetes",
							gitTag,
							filename,
						),
						SHA512: sha512,
						SHA256: sha256,
					},
				})
			}
		}
	}

	imageTarAssets := []eksDistrov1alpha1.Asset{}
	linuxImageArches := []string{"amd64", "arm64"}
	images := []string{
		"kube-apiserver",
		"kube-controller-manager",
		"kube-scheduler",
		"kube-proxy",
		"pause",
	}

	for _, binary := range images {
		assets = append(assets, eksDistrov1alpha1.Asset{
			Name:        fmt.Sprintf("%s-image", binary),
			Type:        "Image",
			Description: fmt.Sprintf("%s container image", binary),
			OS:          "linux",
			Arch:        []string{"amd64", "arm64"},
			Image: &eksDistrov1alpha1.AssetImage{
				URI: fmt.Sprintf("%s/kubernetes/%s:%s",
					r.ContainerImageRepository,
					binary,
					kgv.KubeGitVersion,
				),
			},
		})
		for _, arch := range linuxImageArches {
			filename := path.Join("bin", "linux", arch, fmt.Sprintf("%s.tar", binary))
			sha256, sha512, err := r.ReadK8sShaSums(spec.Channel, filename)
			if err != nil {
				return nil, err
			}
			imageTarAssets = append(imageTarAssets, eksDistrov1alpha1.Asset{
				Name:        filename,
				Type:        "Archive",
				Description: fmt.Sprintf("%s linux/%s OCI image tar", binary, arch),
				OS:          "linux",
				Arch:        []string{arch},
				Archive: &eksDistrov1alpha1.AssetArchive{
					Path: path.Join(
						fmt.Sprintf("kubernetes-%s", spec.Channel),
						"releases",
						fmt.Sprintf("%d", spec.Number),
						"artifacts",
						"kubernetes",
						gitTag,
						filename,
					),
					SHA512: sha512,
					SHA256: sha256,
				},
			})
		}
	}

	assets = append(assets, binaryAssets...)
	assets = append(assets, imageTarAssets...)

	filename := "kubernetes-src.tar.gz"
	sha256, sha512, err := r.ReadK8sShaSums(spec.Channel, filename)
	if err != nil {
		return nil, err
	}
	assets = append(assets, eksDistrov1alpha1.Asset{
		Name:        filename,
		Type:        "Archive",
		Description: "Kubernetes source tarball",
		Archive: &eksDistrov1alpha1.AssetArchive{
			Path: path.Join(
				fmt.Sprintf("kubernetes-%s", spec.Channel),
				"releases",
				fmt.Sprintf("%d", spec.Number),
				"artifacts",
				"kubernetes",
				gitTag,
				filename,
			),
			SHA512: sha512,
			SHA256: sha256,
		},
	})
	component := &eksDistrov1alpha1.Component{
		Name:      "kubernetes",
		GitCommit: kgv.KubeGitCommit,
		GitTag:    gitTag,
		Assets:    assets,
	}

	return component, nil
}