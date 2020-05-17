package vodmodule_stats

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type PodScraper struct {
	Namespace, StatusPath string
	Logger                zerolog.Logger

	podClient v1.PodInterface
}

func (s PodScraper) Scrape() error {
	if err := s.ensure(); err != nil {
		return err
	}

	pods, err := s.fetchPods()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(pods))

	for _, pod := range pods {
		pod := pod
		go func() {
			defer wg.Done()

			logger := s.Logger.With().
				Str("status_url", pod.statusEndpoint).
				Str("name", pod.name).
				Logger()

			resp, err := http.Get(pod.statusEndpoint)
			if err != nil {
				logger.Err(fmt.Errorf("requesting status: %w", err)).Msg("req failed")
				return
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Err(fmt.Errorf("reading response body: %w", err)).Msg("req failed")
				return
			}

			logger.Info().Msgf("status: %s", string(body))
		}()
	}

	wg.Wait()

	return nil
}

type pod struct {
	name           string
	statusEndpoint string
}

func (s PodScraper) fetchPods() ([]pod, error) {
	list, err := s.podClient.List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing pods in namespace %q: %w", s.Namespace, err)
	}

	pods := make([]pod, len(list.Items))
	for i, p := range list.Items {
		pods[i] = pod{
			name: p.Name,
			statusEndpoint: fmt.Sprintf("http://%s/%s", p.Status.PodIP,
				strings.TrimPrefix(s.StatusPath, "/")),
		}
	}

	return pods, nil
}

func (s *PodScraper) ensure() (err error) {
	if s.podClient == nil {
		s.podClient, err = podsClient(s.Namespace)
		if err != nil {
			return fmt.Errorf("creating pods client with namespace %q: %w", s.Namespace, err)
		}
	}

	return nil
}
