package vodmodule_stats

import "encoding/xml"

type CacheStats struct {
	StoreOk      string `xml:"store_ok"`
	StoreBytes   string `xml:"store_bytes"`
	StoreErr     string `xml:"store_err"`
	StoreExists  string `xml:"store_exists"`
	FetchHit     string `xml:"fetch_hit"`
	FetchBytes   string `xml:"fetch_bytes"`
	FetchMiss    string `xml:"fetch_miss"`
	Evicted      string `xml:"evicted"`
	EvictedBytes string `xml:"evicted_bytes"`
	Reset        string `xml:"reset"`
	Entries      string `xml:"entries"`
	DataSize     string `xml:"data_size"`
}

type PerfCounters struct {
	Sum     string `xml:"sum"`
	Count   string `xml:"count"`
	Max     string `xml:"max"`
	MaxTime string `xml:"max_time"`
	MaxPid  string `xml:"max_pid"`
}

type status struct {
	XMLName             xml.Name   `xml:"vod"`
	Version             string     `xml:"version"`
	MetadataCache       CacheStats `xml:"metadata_cache"`
	ResponseCache       CacheStats `xml:"response_cache"`
	MappingCache        CacheStats `xml:"mapping_cache"`
	DrmInfoCache        CacheStats `xml:"drm_info_cache"`
	PerformanceCounters struct {
		FetchCache          PerfCounters `xml:"fetch_cache"`
		StoreCache          PerfCounters `xml:"store_cache"`
		MapPath             PerfCounters `xml:"map_path"`
		ParseMediaSet       PerfCounters `xml:"parse_media_set"`
		GetDrmInfo          PerfCounters `xml:"get_drm_info"`
		OpenFile            PerfCounters `xml:"open_file"`
		AsyncOpenFile       PerfCounters `xml:"async_open_file"`
		ReadFile            PerfCounters `xml:"read_file"`
		AsyncReadFile       PerfCounters `xml:"async_read_file"`
		MediaParse          PerfCounters `xml:"media_parse"`
		BuildManifest       PerfCounters `xml:"build_manifest"`
		InitFrameProcessing PerfCounters `xml:"init_frame_processing"`
		ProcessFrames       PerfCounters `xml:"process_frames"`
		Total               PerfCounters `xml:"total"`
	} `xml:"performance_counters"`
}
