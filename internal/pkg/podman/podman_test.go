package podman_test

import (
	"context"
	"strings"

	"github.com/fromanirh/pack8s/internal/pkg/images"
	"github.com/fromanirh/pack8s/internal/pkg/podman"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/fromanirh/pack8s/iopodman"
)

var _ = Describe("podman", func() {
	Context("create new handler", func() {
		It("Should create with default socket without any error", func() {
			ctx := context.Background()

			_, err := podman.NewHandle(ctx, "")
			Expect(err).To(BeNil())
		})

		It("Should create with user defined socket without any error", func() {
			ctx := context.Background()

			_, err := podman.NewHandle(ctx, "unix:/run/podman/io.podman")
			Expect(err).To(BeNil())
		})
	})

	Context("create volume", func() {
		It("Should create volume", func() {
			ctx := context.Background()

			handler, err := podman.NewHandle(ctx, "")
			Expect(err).To(BeNil())

			_, err = handler.CreateNamedVolume("pack8s-test")
			Expect(err).To(BeNil())

			volumes, err := handler.GetAllVolumes()
			Expect(err).To(BeNil())

			found := false
			for _, volume := range volumes {
				if volume.Name == "pack8s-test" {
					found = true
				}
			}

			Expect(found).To(Equal(true))

			volumesToRemove := []iopodman.Volume{
				{
					Name: "pack8s-test",
				},
			}
			err = handler.RemoveVolumes(volumesToRemove)
			Expect(err).To(BeNil())
		})
	})

	Context("find prefixed volumes", func() {
		It("Should get prefixed volume", func() {
			ctx := context.Background()

			handler, err := podman.NewHandle(ctx, "")
			Expect(err).To(BeNil())

			_, err = handler.CreateNamedVolume("pack8s-1-test")
			Expect(err).To(BeNil())

			_, err = handler.CreateNamedVolume("pack8s-2-test")
			Expect(err).To(BeNil())

			_, err = handler.CreateNamedVolume("some-pack-test")
			Expect(err).To(BeNil())

			volumes, err := handler.GetPrefixedVolumes("pack8s")
			Expect(err).To(BeNil())

			Expect(len(volumes)).To(Equal(2))

			volumesToRemove := []iopodman.Volume{
				{
					Name: "pack8s-1-test",
				}, {
					Name: "pack8s-2-test",
				}, {
					Name: "some-pack-test",
				},
			}
			err = handler.RemoveVolumes(volumesToRemove)
			Expect(err).To(BeNil())
		})
	})

	Context("remove volumes", func() {
		It("Should remove volume", func() {
			ctx := context.Background()

			handler, err := podman.NewHandle(ctx, "")
			Expect(err).To(BeNil())

			_, err = handler.CreateNamedVolume("pack8s-test")
			Expect(err).To(BeNil())
			_, err = handler.CreateNamedVolume("pack8s-test-do-not-delete")
			Expect(err).To(BeNil())

			volumesToRemove := []iopodman.Volume{
				{
					Name: "pack8s-test",
				},
			}
			err = handler.RemoveVolumes(volumesToRemove)
			Expect(err).To(BeNil())

			volumes, err := handler.GetAllVolumes()
			Expect(err).To(BeNil())

			found := false
			deletedFound := false
			for _, volume := range volumes {
				if volume.Name == "pack8s-test-do-not-delete" {
					found = true
				}

				if volume.Name == "pack8s-test" {
					deletedFound = true
				}
			}

			Expect(found).To(Equal(true))
			Expect(deletedFound).To(Equal(false))

			volumesToRemove = []iopodman.Volume{
				{
					Name: "pack8s-test-do-not-delete",
				},
			}
			err = handler.RemoveVolumes(volumesToRemove)
			Expect(err).To(BeNil())
		})
	})

	Context("create container", func() {
		It("Should create container", func() {
			ctx := context.Background()

			handler, err := podman.NewHandle(ctx, "")
			Expect(err).To(BeNil())

			name := "pack8s-test"
			id, err := handler.CreateContainer(iopodman.Create{
				Args: []string{images.DockerRegistryImage},
				Name: &name,
			})
			Expect(err).To(BeNil())

			container, err := handler.FindPrefixedContainer(name)
			Expect(err).To(BeNil())
			Expect(container.Id).To(Equal(id))

			_, err = handler.RemoveContainer(iopodman.Container{Id: id}, true, true)
			Expect(err).To(BeNil())
		})
	})

	Context("find container", func() {
		It("Should find 1 container", func() {
			ctx := context.Background()

			handler, err := podman.NewHandle(ctx, "")
			Expect(err).To(BeNil())
			name1 := "pack8s-test"
			id1, err := handler.CreateContainer(iopodman.Create{
				Args: []string{images.DockerRegistryImage},
				Name: &name1,
			})
			Expect(err).To(BeNil())

			name2 := "test-1"
			id2, err := handler.CreateContainer(iopodman.Create{
				Args: []string{images.DockerRegistryImage},
				Name: &name2,
			})
			Expect(err).To(BeNil())

			container, err := handler.FindPrefixedContainer(name1)
			Expect(err).To(BeNil())
			Expect(container.Id).To(Equal(id1))

			_, err = handler.RemoveContainer(iopodman.Container{Id: id1}, true, true)
			Expect(err).To(BeNil())

			_, err = handler.RemoveContainer(iopodman.Container{Id: id2}, true, true)
			Expect(err).To(BeNil())
		})

		It("Should throw error when 2 containers match prefix", func() {
			ctx := context.Background()

			handler, err := podman.NewHandle(ctx, "")
			Expect(err).To(BeNil())
			name1 := "pack8s-test"
			id1, err := handler.CreateContainer(iopodman.Create{
				Args: []string{images.DockerRegistryImage},
				Name: &name1,
			})
			Expect(err).To(BeNil())

			name2 := "pack8s-test-1"
			id2, err := handler.CreateContainer(iopodman.Create{
				Args: []string{images.DockerRegistryImage},
				Name: &name2,
			})
			Expect(err).To(BeNil())

			_, err = handler.FindPrefixedContainer(name1)
			Expect(err).NotTo(BeNil())

			_, err = handler.RemoveContainer(iopodman.Container{Id: id1}, true, true)
			Expect(err).To(BeNil())

			_, err = handler.RemoveContainer(iopodman.Container{Id: id2}, true, true)
			Expect(err).To(BeNil())
		})
	})

	Context("remove container", func() {
		It("Should remove container", func() {
			ctx := context.Background()

			handler, err := podman.NewHandle(ctx, "")
			Expect(err).To(BeNil())

			name := "pack8s-test"
			id, err := handler.CreateContainer(iopodman.Create{
				Args: []string{images.DockerRegistryImage},
				Name: &name,
			})
			Expect(err).To(BeNil())

			container, err := handler.FindPrefixedContainer(name)
			Expect(err).To(BeNil())
			Expect(container.Id).To(Equal(id))

			_, err = handler.RemoveContainer(iopodman.Container{Id: id}, true, true)
			Expect(err).To(BeNil())

			_, err = handler.FindPrefixedContainer(name)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("pull image", func() {
		It("Should pull image", func() {
			ctx := context.Background()

			handler, err := podman.NewHandle(ctx, "")
			Expect(err).To(BeNil())

			err = handler.PullImage(images.DockerRegistryImage)
			Expect(err).To(BeNil())

			images, err := handler.ListImages()

			found := false

			for _, image := range images {
				for _, tag := range image.RepoTags {
					if strings.Contains(tag, "docker.io/library/registry") {
						found = true
					}
				}
			}

			Expect(found).To(Equal(true))
		})
	})
})