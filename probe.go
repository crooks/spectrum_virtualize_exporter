// All currently supported probes
//
// Copyright (C) 2020  Christian Svensson
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/alecthomas/units"
	"github.com/prometheus/client_golang/prometheus"
)

func strBool64(s string) (bool64 float64) {
	sLower := strings.ToLower(s)
	if sLower == "y" || sLower == "yes" {
		bool64 = 1.0
	}
	return
}

func probeNodeStats(c SpectrumHTTP, registry *prometheus.Registry) bool {
	var (
		mCmpCPU = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_node_compression_usage_ratio",
				Help: "Current ratio of allocated CPU for compresion",
			},
			[]string{"id"},
		)
		mSysCPU = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_node_system_usage_ratio",
				Help: "Current ratio of allocated CPU for system",
			},
			[]string{"id"},
		)
		mCacheWrite = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_node_write_cache_usage_ratio",
				Help: "Ratio of the write cache usage for the node",
			},
			[]string{"id"},
		)
		mCacheTotal = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_node_total_cache_usage_ratio",
				Help: "Total percentage for both the write and read cache usage for the node",
			},
			[]string{"id"},
		)
		mFcBytes = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_node_fc_bps",
				Help: "Current bytes-per-second being transferred over Fibre Channel",
			},
			[]string{"id"},
		)
		mFcIO = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_node_fc_iops",
				Help: "Current I/O-per-second being transferred over Fibre Channel",
			},
			[]string{"id"},
		)
		mISCSIBytes = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_node_iscsi_bps",
				Help: "Current bytes-per-second being transferred over iSCSI",
			},
			[]string{"id"},
		)
		mISCSIIO = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_node_iscsi_iops",
				Help: "Current I/O-per-second being transferred over iSCSI",
			},
			[]string{"id"},
		)
		mSASBytes = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_node_sas_bps",
				Help: "Current bytes-per-second being transferred over backend SAS",
			},
			[]string{"id"},
		)
		mSASIO = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_node_sas_iops",
				Help: "Current I/O-per-second being transferred over backend SAS",
			},
			[]string{"id"},
		)
	)

	registry.MustRegister(mSysCPU)
	registry.MustRegister(mCmpCPU)
	registry.MustRegister(mCacheWrite)
	registry.MustRegister(mCacheTotal)
	registry.MustRegister(mFcBytes)
	registry.MustRegister(mFcIO)
	registry.MustRegister(mISCSIBytes)
	registry.MustRegister(mISCSIIO)
	registry.MustRegister(mSASBytes)
	registry.MustRegister(mSASIO)

	type nodeStat struct {
		NodeID      string `json:"node_id"`
		StatName    string `json:"stat_name"`
		StatCurrent int    `json:"stat_current,string"`
	}
	var st []nodeStat

	if err := c.Get("rest/lsnodecanisterstats", "", &st); err != nil {
		log.Printf("Error: %v", err)
		return false
	}

	for _, s := range st {
		if s.StatName == "compression_cpu_pc" {
			mCmpCPU.WithLabelValues(s.NodeID).Set(float64(s.StatCurrent) / 100.0)
		} else if s.StatName == "cpu_pc" {
			mSysCPU.WithLabelValues(s.NodeID).Set(float64(s.StatCurrent) / 100.0)
		} else if s.StatName == "fc_mb" {
			mFcBytes.WithLabelValues(s.NodeID).Set(float64(s.StatCurrent) * 1024 * 1024)
		} else if s.StatName == "fc_io" {
			mFcIO.WithLabelValues(s.NodeID).Set(float64(s.StatCurrent))
		} else if s.StatName == "iscsi_mb" {
			mISCSIBytes.WithLabelValues(s.NodeID).Set(float64(s.StatCurrent) * 1024 * 1024)
		} else if s.StatName == "iscsi_io" {
			mISCSIIO.WithLabelValues(s.NodeID).Set(float64(s.StatCurrent))
		} else if s.StatName == "sas_mb" {
			mSASBytes.WithLabelValues(s.NodeID).Set(float64(s.StatCurrent) * 1024 * 1024)
		} else if s.StatName == "sas_io" {
			mSASIO.WithLabelValues(s.NodeID).Set(float64(s.StatCurrent))
		} else if s.StatName == "write_cache_pc" {
			mCacheWrite.WithLabelValues(s.NodeID).Set(float64(s.StatCurrent) / 100.0)
		} else if s.StatName == "total_cache_pc" {
			mCacheTotal.WithLabelValues(s.NodeID).Set(float64(s.StatCurrent) / 100.0)
		}
	}
	return true
}

func probeEnclosureStats(c SpectrumHTTP, registry *prometheus.Registry) bool {
	var (
		mPower = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_power_watts",
				Help: "Current power draw of enclosure in watts",
			},
			[]string{"enclosure"},
		)
		mTemp = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_temperature",
				Help: "Current enclosure temperature in celsius",
			},
			[]string{"enclosure"},
		)
	)

	registry.MustRegister(mPower)
	registry.MustRegister(mTemp)

	type enclosureStats struct {
		EnclosureID string `json:"enclosure_id"`
		StatName    string `json:"stat_name"`
		StatCurrent int    `json:"stat_current,string"`
	}
	var st []enclosureStats

	if err := c.Get("rest/lsenclosurestats", "", &st); err != nil {
		log.Printf("Error: %v", err)
		return false
	}

	for _, s := range st {
		if s.StatName == "power_w" {
			mPower.WithLabelValues(s.EnclosureID).Set(float64(s.StatCurrent))
		} else if s.StatName == "temp_c" {
			mTemp.WithLabelValues(s.EnclosureID).Set(float64(s.StatCurrent))
		}
	}
	return true
}

func probeDrives(c SpectrumHTTP, registry *prometheus.Registry) bool {
	labels := []string{"enclosure", "slot_id", "id"}
	var (
		mStatus = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_drive_status",
				Help: "Status of drive",
			},
			append(labels, "status"),
		)
	)

	registry.MustRegister(mStatus)

	type drive struct {
		ID          string
		Status      string
		Use         string
		Capacity    string
		SlotID      string `json:"slot_id"`
		MdiskID     string `json:"mdisk_id"`
		MdiskName   string `json:"mdisk_name"`
		EnclosureID string `json:"enclosure_id"`
	}
	var st []drive

	if err := c.Get("rest/lsdrive", "", &st); err != nil {
		log.Printf("Error: %v", err)
		return false
	}

	for _, s := range st {
		var son, soff, sdeg float64
		if s.Status == "online" {
			son = 1.0
		} else if s.Status == "offline" {
			soff = 1.0
		} else if s.Status == "degraded" {
			sdeg = 1.0
		}
		mStatus.WithLabelValues(s.EnclosureID, s.SlotID, s.ID, "online").Set(float64(son))
		mStatus.WithLabelValues(s.EnclosureID, s.SlotID, s.ID, "offline").Set(float64(soff))
		mStatus.WithLabelValues(s.EnclosureID, s.SlotID, s.ID, "degraded").Set(float64(sdeg))
	}
	return true
}

func probeEnclosurePSUs(c SpectrumHTTP, registry *prometheus.Registry) bool {
	labels := []string{"enclosure", "id"}
	var (
		mStatus = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_psu_status",
				Help: "Status of PSU",
			},
			append(labels, "status"),
		)
	)

	registry.MustRegister(mStatus)

	type psu struct {
		Status      string
		PSUID       string `json:"psu_id"`
		EnclosureID string `json:"enclosure_id"`
	}
	var st []psu

	if err := c.Get("rest/lsenclosurepsu", "", &st); err != nil {
		log.Printf("Error: %v", err)
		return false
	}

	for _, s := range st {
		var son, soff, sdeg float64
		if s.Status == "online" {
			son = 1.0
		} else if s.Status == "offline" {
			soff = 1.0
		} else if s.Status == "degraded" {
			sdeg = 1.0
		}
		mStatus.WithLabelValues(s.EnclosureID, s.PSUID, "online").Set(float64(son))
		mStatus.WithLabelValues(s.EnclosureID, s.PSUID, "offline").Set(float64(soff))
		mStatus.WithLabelValues(s.EnclosureID, s.PSUID, "degraded").Set(float64(sdeg))
	}
	return true
}

func probePool(c SpectrumHTTP, registry *prometheus.Registry) bool {
	labels := []string{"id", "name"}
	var (
		mStatus = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_pool_status",
				Help: "Status of pool",
			},
			append(labels, "status"),
		)
		mVdiskCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "spectrum_pool_volume_count", Help: "Number of volumes associated with pool"}, labels)
		mCapacity   = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "spectrum_pool_capacity_bytes", Help: "Capacity of pool in bytes"}, labels)
		mFree       = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "spectrum_pool_free_bytes", Help: "Free bytes in pool"}, labels)
		mUsed       = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "spectrum_pool_used_bytes", Help: "Used bytes in pool"}, labels)
		mReal       = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "spectrum_pool_real_bytes", Help: "Real capacity bytes in pool"}, labels)
	)

	registry.MustRegister(mStatus)
	registry.MustRegister(mVdiskCount)
	registry.MustRegister(mCapacity)
	registry.MustRegister(mFree)
	registry.MustRegister(mUsed)
	registry.MustRegister(mReal)

	type pool struct {
		ID                  string
		Status              string
		Name                string
		VdiskCount          int `json:"vdisk_count,string"`
		Capacity            string
		FreeCapacity        string `json:"free_capacity"`
		VirtualCapacity     string `json:"virtual_capacity"`
		UsedCapacity        string `json:"used_capacity"`
		RealCapacity        string `json:"real_capacity"`
		ReclaimableCapacity string `json:"reclaimable_capacity"`
	}
	var st []pool

	if err := c.Get("rest/lsmdiskgrp", "", &st); err != nil {
		log.Printf("Error: %v", err)
		return false
	}

	for _, s := range st {
		var son, soff float64
		if s.Status == "online" {
			son = 1.0
		} else if s.Status == "offline" {
			soff = 1.0
		}
		mStatus.WithLabelValues(s.ID, s.Name, "online").Set(float64(son))
		mStatus.WithLabelValues(s.ID, s.Name, "offline").Set(float64(soff))

		mVdiskCount.WithLabelValues(s.ID, s.Name).Set(float64(s.VdiskCount))

		free, err := units.ParseBase2Bytes(s.FreeCapacity)
		if err != nil {
			log.Printf("Failed to parse %q: %v", s.FreeCapacity, err)
		} else {
			mFree.WithLabelValues(s.ID, s.Name).Set(float64(free))
		}

		capacity, err := units.ParseBase2Bytes(s.Capacity)
		if err != nil {
			log.Printf("Failed to parse %q: %v", s.Capacity, err)
		} else {
			mCapacity.WithLabelValues(s.ID, s.Name).Set(float64(capacity))
		}

		used, err := units.ParseBase2Bytes(s.UsedCapacity)
		if err != nil {
			log.Printf("Failed to parse %q: %v", s.UsedCapacity, err)
		} else {
			mUsed.WithLabelValues(s.ID, s.Name).Set(float64(used))
		}

		real, err := units.ParseBase2Bytes(s.RealCapacity)
		if err != nil {
			log.Printf("Failed to parse %q: %v", s.RealCapacity, err)
		} else {
			mReal.WithLabelValues(s.ID, s.Name).Set(float64(real))
		}
	}
	return true
}

func probeHost(c SpectrumHTTP, registry *prometheus.Registry) bool {
	// TODO
	return true
}

func probeFCPorts(c SpectrumHTTP, registry *prometheus.Registry) bool {
	labels := []string{"node_id", "adapter_location", "adapter_port_id"}
	var (
		mStatus = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_fc_port_status",
				Help: "Status of Fibre Channel port",
			},
			append(labels, "wwpn", "status"),
		)
		mSpeed = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_fc_port_speed_bps",
				Help: "Operational speed of port in bits per second",
			},
			append(labels),
		)
	)

	registry.MustRegister(mStatus)
	registry.MustRegister(mSpeed)

	type fcPort struct {
		Type            string
		PortSpeed       string `json:"port_speed"`
		Status          string
		WWPN            string
		NodeID          string `json:"node_id"`
		AdapterLocation string `json:"adapter_location"`
		AdapterPortIID  string `json:"adapter_port_id"`
	}
	var st []fcPort

	if err := c.Get("rest/lsportfc", "", &st); err != nil {
		log.Printf("Error: %v", err)
		return false
	}

	for _, s := range st {
		var online, inunc, inc float64
		if s.Status == "active" {
			online = 1.0
		} else if s.Status == "inactive_unconfigured" {
			inunc = 1.0
		} else if s.Status == "inactive_configured" {
			inc = 1.0
		}
		mStatus.WithLabelValues(s.NodeID, s.AdapterLocation, s.AdapterPortIID, s.WWPN, "active").Set(float64(online))
		mStatus.WithLabelValues(s.NodeID, s.AdapterLocation, s.AdapterPortIID, s.WWPN, "inactive_unconfigured").Set(float64(inunc))
		mStatus.WithLabelValues(s.NodeID, s.AdapterLocation, s.AdapterPortIID, s.WWPN, "inactive_configured").Set(float64(inc))

		ps := 0
		if pss := strings.TrimSuffix(s.PortSpeed, "Gb"); pss != s.PortSpeed {
			x, err := strconv.Atoi(pss)
			if err == nil {
				ps = x * 1000 * 1000 * 1000
			}
		}
		mSpeed.WithLabelValues(s.NodeID, s.AdapterLocation, s.AdapterPortIID).Set(float64(ps))
	}
	return true
}

func probeIPPorts(c SpectrumHTTP, registry *prometheus.Registry) bool {
	labels := []string{"node_id", "adapter_location", "adapter_port_id"}
	var (
		mState = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_ip_port_state",
				Help: "Configuration state of Ethernet/IP port",
			},
			append(labels, "mac", "state"),
		)
		mActive = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_ip_port_link_active",
				Help: "Whether link is active",
			},
			append(labels, "mac"),
		)
		mSpeed = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_ip_port_speed_bps",
				Help: "Operational speed of port in bits per second",
			},
			append(labels),
		)
	)

	registry.MustRegister(mState)
	registry.MustRegister(mActive)
	registry.MustRegister(mSpeed)

	type ipPort struct {
		Speed           string
		State           string
		LinkState       string `json:"link_state"`
		MAC             string
		NodeID          string `json:"node_id"`
		AdapterLocation string `json:"adapter_location"`
		AdapterPortIID  string `json:"adapter_port_id"`
	}
	var st []ipPort

	if err := c.Get("rest/lsportip", "", &st); err != nil {
		log.Printf("Error: %v", err)
		return false
	}

	for _, s := range st {
		var con, uncon, mgmt float64
		if s.State == "configured" {
			con = 1.0
		} else if s.State == "unconfigured" {
			uncon = 1.0
		} else if s.State == "management_only" {
			mgmt = 1.0
		}
		mState.WithLabelValues(s.NodeID, s.AdapterLocation, s.AdapterPortIID, s.MAC, "configured").Set(float64(con))
		mState.WithLabelValues(s.NodeID, s.AdapterLocation, s.AdapterPortIID, s.MAC, "unconfigured").Set(float64(uncon))
		mState.WithLabelValues(s.NodeID, s.AdapterLocation, s.AdapterPortIID, s.MAC, "management_only").Set(float64(mgmt))

		active := 0
		if s.LinkState == "active" {
			active = 1
		}
		mActive.WithLabelValues(s.NodeID, s.AdapterLocation, s.AdapterPortIID, s.MAC).Set(float64(active))

		ps := 0
		if pss := strings.TrimSuffix(s.Speed, "Gb/s"); pss != s.Speed {
			x, err := strconv.Atoi(pss)
			if err == nil {
				ps = x * 1000 * 1000 * 1000
			}
		}
		if pss := strings.TrimSuffix(s.Speed, "Mb/s"); pss != s.Speed {
			x, err := strconv.Atoi(pss)
			if err == nil {
				ps = x * 1000 * 1000
			}
		}
		mSpeed.WithLabelValues(s.NodeID, s.AdapterLocation, s.AdapterPortIID).Set(float64(ps))
	}
	return true
}

func probeRcConsistGrp(c SpectrumHTTP, registry *prometheus.Registry) bool {
	type iorcgrp struct {
		ID         string
		Name       string
		MasterID   string `json:"master_cluster_id"`
		MasterName string `json:"master_cluster_name"`
		AuxID      string `json:"aux_cluster_id"`
		AuxName    string `json:"aux_cluster_name"`
		Primary    string `json:"primary"`
		State      string
		RelCount   string `json:"relationship_count"`
		CopyType   string `json:"copy_type"`
		CycleMode  string `json:"cycling_mode"`
		FreezeTime string `json:"freeze_time"`
	}
	labels := []string{"name"}
	var (
		mState = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_rc_consistency_state",
				Help: "Remote Copy consistency status",
			},
			append(labels, "master", "aux"),
		)
	)

	registry.MustRegister(mState)

	var st []iorcgrp
	if err := c.Get("rest/lsrcconsistgrp", "", &st); err != nil {
		log.Printf("Error: %v", err)
		return false
	}

	for _, s := range st {
		var state float64
		// Leaving scope to explicitly return states (1-9)
		if s.State == "consistent_synchronized" {
			state = 10.0
		} else {
			state = 0.0
		}
		mState.WithLabelValues(s.Name, s.MasterName, s.AuxName).Set(state)
	}
	return true
}

func probeIOgrpsDetail(c SpectrumHTTP, registry *prometheus.Registry, ids []string) bool {
	labels := []string{"id", "name"}
	var (
		iFlashTotal = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_iogrp_mem_flash_total",
				Help: "Total memory allocated to FlashCopy",
			},
			append(labels),
		)
		iFlashFree = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_iogrp_mem_flash_free",
				Help: "Unused memory allocated to FlashCopy",
			},
			append(labels),
		)
		iFlashMax = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_iogrp_mem_flash_max",
				Help: "Maximum memory that can be allocated to FlashCopy",
			},
			append(labels),
		)
		iRaidTotal = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_iogrp_mem_raid_total",
				Help: "Total memory allocated to RAID",
			},
			append(labels),
		)
		iRaidFree = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_iogrp_mem_raid_free",
				Help: "Unused memory allocated to RAID",
			},
			append(labels),
		)
		iMaintenance = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "spectrum_iogrp_maintenance",
				Help: "Is the I/O group in maintenance mode",
			},
			append(labels),
		)
	)

	registry.MustRegister(iFlashTotal)
	registry.MustRegister(iFlashFree)
	registry.MustRegister(iFlashMax)
	registry.MustRegister(iRaidTotal)
	registry.MustRegister(iRaidFree)
	registry.MustRegister(iMaintenance)

	type ioGrp struct {
		ID                 string
		Name               string
		NodeCount          string `json:"node_count"`
		VdiskCount         string `json:"vdisk_count"`
		HostCount          string `json:"host_count"`
		FlashCopyTotalMem  string `json:"flash_copy_total_memory"`
		FlashCopyFreeMem   string `json:"flash_copy_free_memory"`
		RemoteCopyTotalMem string `json:"remote_copy_total_memory"`
		RemoteCopyFreeMem  string `json:"remote_copy_free_memory"`
		MirrorTotalMem     string `json:"mirroring_total_memory"`
		MirrorFreeMem      string `json:"mirroring_free_memory"`
		RaidTotalMem       string `json:"raid_total_memory"`
		RaidFreeMem        string `json:"raid_free_memory"`
		Maintenance        string `json:"maintenance"`
		Compression        string `json:"compression_active"`
		AccessVdiskCount   string `json:"accessible_vdisk_count"`
		CompressionSupport string `json:"compression_supported"`
		MaxEnclosures      string `json:"max_enclosures"`
		Encryption         string `json:"encryption_supported"`
		FlashCopyMaxMem    string `json:"flash_copy_maximum_memory"`
		SiteID             string `json:"site_id"`
		SiteName           string `json:"site_name"`
		FCTargetPortMode   string
		CompressTotalMem   string `json:"compression_total_memory"`
		Duplication        string `json:"deduplication_supported"`
		Deduplication      string `json:"deduplication_active"`
		NQN                string
	}

	var st []ioGrp
	for _, id := range ids {
		var stn ioGrp
		url := fmt.Sprintf("rest/lsiogrp/%s", id)
		if err := c.Get(url, "", &stn); err != nil {
			log.Printf("Error in Get: %v", err)
			return false
		}
		st = append(st, stn)
	}
	for _, s := range st {
		ftot, err := units.ParseBase2Bytes(s.FlashCopyTotalMem)
		if err != nil {
			log.Printf("Failed to parse %q: %v", s.FlashCopyTotalMem, err)
		} else {
			iFlashTotal.WithLabelValues(s.ID, s.Name).Set(float64(ftot))
		}
		ffree, err := units.ParseBase2Bytes(s.FlashCopyFreeMem)
		if err != nil {
			log.Printf("Failed to parse %q: %v", s.FlashCopyFreeMem, err)
		} else {
			iFlashFree.WithLabelValues(s.ID, s.Name).Set(float64(ffree))
		}
		fmax, err := units.ParseBase2Bytes(s.FlashCopyMaxMem)
		if err != nil {
			log.Printf("Failed to parse %q: %v", s.FlashCopyMaxMem, err)
		} else {
			iFlashMax.WithLabelValues(s.ID, s.Name).Set(float64(fmax))
		}
		rtot, err := units.ParseBase2Bytes(s.RaidTotalMem)
		if err != nil {
			log.Printf("Failed to parse %q: %v", s.RaidTotalMem, err)
		} else {
			iRaidTotal.WithLabelValues(s.ID, s.Name).Set(float64(rtot))
		}
		rfree, err := units.ParseBase2Bytes(s.RaidFreeMem)
		if err != nil {
			log.Printf("Failed to parse %q: %v", s.RaidFreeMem, err)
		} else {
			iRaidFree.WithLabelValues(s.ID, s.Name).Set(float64(rfree))
		}
		// Maintenance Mode
		iMaintenance.WithLabelValues(s.ID, s.Name).Set(strBool64(s.Maintenance))
	}
	return true
}

func getIOgrps(c SpectrumHTTP, registry *prometheus.Registry) (ids []string, err error) {
	type iogrp_concise struct {
		ID         string
		Name       string
		NodeCount  string `json:"node_count"`
		VdiskCount string `json:"vdisk_count"`
		HostCount  string `json:"host_count"`
		SiteID     string `json:"site_id"`
		SiteName   string `json:"site_name"`
	}
	var st []iogrp_concise
	if err = c.Get("rest/lsiogrp", "", &st); err != nil {
		log.Printf("Error: %v", err)
		return
	}

	for _, s := range st {
		// nc and vc are integer conversions of NodeCount and VdiskCount.  We
		// want to excude IDs where either of these are zero.
		nc, err := strconv.Atoi(s.NodeCount)
		if err != nil {
			log.Printf("Error: %v", err)
		}
		vc, err := strconv.Atoi(s.VdiskCount)
		if err != nil {
			log.Printf("Error: %v", err)
		}
		if nc > 0 || vc > 0 {
			ids = append(ids, s.ID)
		}
	}
	return
}

func probe(ctx context.Context, target string, registry *prometheus.Registry, hc *http.Client) (bool, error) {
	tgt, err := url.Parse(target)
	if err != nil {
		return false, fmt.Errorf("url.Parse failed: %v", err)
	}

	if tgt.Scheme != "https" && tgt.Scheme != "http" {
		return false, fmt.Errorf("Unsupported scheme %q", tgt.Scheme)
	}

	// Filter anything else than scheme and hostname
	u := url.URL{
		Scheme: tgt.Scheme,
		Host:   tgt.Host,
	}
	c, err := newSpectrumClient(ctx, u, hc)
	if err != nil {
		return false, err
	}

	// TODO: Make parallel
	ioGrpIds, err := getIOgrps(c, registry)
	if err != nil {
		return false, fmt.Errorf("getIOgrps() returned: %v", err)
	}
	success := probeEnclosureStats(c, registry) &&
		probeEnclosurePSUs(c, registry) &&
		probePool(c, registry) &&
		probeDrives(c, registry) &&
		probeNodeStats(c, registry) &&
		probeHost(c, registry) &&
		probeFCPorts(c, registry) &&
		probeIPPorts(c, registry) &&
		probeRcConsistGrp(c, registry) &&
		probeIOgrpsDetail(c, registry, ioGrpIds)

	return success, nil
}
