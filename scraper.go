package vodmodule_stats

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type PodScraper struct {
	Namespace, StatusPath string
	Logger                zerolog.Logger
	Client                *http.Client

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

			resp, err := s.Client.Get(pod.statusEndpoint)
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

			var s status

			err = xml.Unmarshal(body, &s)
			if err != nil {
				logger.Err(fmt.Errorf("unmarshaling xml response: %w", err)).Msg("parse failed")
				return
			}

			cacheStatses := map[string]CacheStats{
				"c_metadata": s.MetadataCache,
				"c_response": s.ResponseCache,
				"c_mapping":  s.MappingCache,
				"c_drm_info": s.DrmInfoCache,
			}

			for k, cs := range cacheStatses {
				logger = logger.With().
					Int(fmt.Sprintf("%s_store_ok", k), mustInt(cs.StoreOk)).
					Int(fmt.Sprintf("%s_store_bytes", k), mustInt(cs.StoreBytes)).
					Int(fmt.Sprintf("%s_store_err", k), mustInt(cs.StoreErr)).
					Int(fmt.Sprintf("%s_store_exists", k), mustInt(cs.StoreExists)).
					Int(fmt.Sprintf("%s_fetch_hit", k), mustInt(cs.FetchHit)).
					Int(fmt.Sprintf("%s_fetch_bytes", k), mustInt(cs.FetchBytes)).
					Int(fmt.Sprintf("%s_fetch_miss", k), mustInt(cs.FetchMiss)).
					Int(fmt.Sprintf("%s_evicted", k), mustInt(cs.Evicted)).
					Int(fmt.Sprintf("%s_evicted_bytes", k), mustInt(cs.EvictedBytes)).
					Int(fmt.Sprintf("%s_reset", k), mustInt(cs.Reset)).
					Int(fmt.Sprintf("%s_entries", k), mustInt(cs.Entries)).
					Int(fmt.Sprintf("%s_data_size", k), mustInt(cs.DataSize)).
					Logger()
			}

			counters := s.PerformanceCounters

			performanceCounters := map[string]PerfCounters{
				"pc_fetch_cache":     counters.FetchCache,
				"pc_store_cache":     counters.StoreCache,
				"pc_map_path":        counters.MapPath,
				"pc_parse_media_set": counters.ParseMediaSet,
				"pc_get_drm_info":    counters.GetDrmInfo,
				"pc_open_file":       counters.OpenFile,
				"pc_async_open_file": counters.AsyncOpenFile,
				"pc_read_file":       counters.ReadFile,
				"pc_async_read_file": counters.AsyncReadFile,
				"pc_media_parse":     counters.MediaParse,
				"pc_build_manifest":  counters.BuildManifest,
				"pc_init_frame_proc": counters.InitFrameProcessing,
				"pc_proc_frames":     counters.ProcessFrames,
				"pc_total":           counters.Total,
			}

			for k, pc := range performanceCounters {
				logger = logger.With().
					Int(fmt.Sprintf("%s_sum", k), mustInt(pc.Sum)).
					Int(fmt.Sprintf("%s_count", k), mustInt(pc.Count)).
					Int(fmt.Sprintf("%s_max", k), mustInt(pc.Max)).
					Int(fmt.Sprintf("%s_max_time", k), mustInt(pc.MaxTime)).
					Int(fmt.Sprintf("%s_max_pid", k), mustInt(pc.MaxPid)).
					Logger()
			}

			logger.Info().Msg("successful scrape")
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

	if s.Client == nil {
		s.Client = &http.Client{Timeout: 10 * time.Second}
	}

	return nil
}
