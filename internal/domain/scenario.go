package domain

// Scenario 剧本
type Scenario struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	Anomaly          *AnomalyProfile   `json:"anomaly"`
	MorningScenes    []*MorningScene   `json:"morning_scenes"`
	Briefing         *Briefing         `json:"briefing"`
	OptionalGoals    []*OptionalGoal   `json:"optional_goals"`
	Scenes           map[string]*Scene `json:"scenes"`
	StartingSceneID  string            `json:"starting_scene_id"`
	Encounter        *Encounter        `json:"encounter"`
	Aftermath        *Aftermath        `json:"aftermath"`
	Rewards          *Rewards          `json:"rewards"`
}

// AnomalyProfile 异常体档案
type AnomalyProfile struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	History       string         `json:"history"`
	Focus         *Focus         `json:"focus"`
	Domain        *Domain        `json:"domain"`
	Appearance    string         `json:"appearance"`
	Impulse       string         `json:"impulse"`
	CurrentStatus string         `json:"current_status"`
	ChaosEffects  []*ChaosEffect `json:"chaos_effects"`
}

// Focus 焦点
type Focus struct {
	Emotion string `json:"emotion"`
	Subject string `json:"subject"`
}

// Domain 领域
type Domain struct {
	Location    string `json:"location"`
	Description string `json:"description"`
}

// ChaosEffect 混沌效应
type ChaosEffect struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Cost        int    `json:"cost"`
	Description string `json:"description"`
	Effect      string `json:"effect"`
}

// MorningScene 晨会场景
type MorningScene struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// Briefing 任务简报
type Briefing struct {
	Summary     string   `json:"summary"`
	Objectives  []string `json:"objectives"`
	Warnings    []string `json:"warnings"`
}

// OptionalGoal 可选目标
type OptionalGoal struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Reward      int    `json:"reward"`
}

// Scene 场景
type Scene struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	NPCs        []*NPC            `json:"npcs"`
	Clues       []*Clue           `json:"clues"`
	Events      []*Event          `json:"events"`
	Connections []string          `json:"connections"`
	State       map[string]interface{} `json:"state"`
}

// NPC 非玩家角色
type NPC struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Personality string   `json:"personality"`
	Dialogues   []string `json:"dialogues"`
	State       string   `json:"state"`
}

// Clue 线索
type Clue struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Requirements []string `json:"requirements"`
	Unlocks      []string `json:"unlocks"`
}

// Event 事件
type Event struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Trigger     string   `json:"trigger"`
	Effect      string   `json:"effect"`
}

// Encounter 遭遇
type Encounter struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Phases      []*Phase `json:"phases"`
}

// Phase 遭遇阶段
type Phase struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Actions     []string `json:"actions"`
}

// Aftermath 余波
type Aftermath struct {
	Captured    string `json:"captured"`
	Neutralized string `json:"neutralized"`
	Escaped     string `json:"escaped"`
}

// Rewards 奖励
type Rewards struct {
	Commendations int      `json:"commendations"`
	Claimables    []string `json:"claimables"`
}
