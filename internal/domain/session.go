package domain

import "time"

// GameSession 游戏会话
type GameSession struct {
	ID         string     `json:"id"`
	AgentID    string     `json:"agent_id"`
	ScenarioID string     `json:"scenario_id"`
	Phase      GamePhase  `json:"phase"`
	State      *GameState `json:"state"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// GamePhase 游戏阶段
type GamePhase string

const (
	PhaseMorning       GamePhase = "morning"
	PhaseInvestigation GamePhase = "investigation"
	PhaseEncounter     GamePhase = "encounter"
	PhaseAftermath     GamePhase = "aftermath"
)

// GameState 游戏状态
type GameState struct {
	CurrentSceneID    string               `json:"current_scene_id"`
	VisitedScenes     map[string]bool      `json:"visited_scenes"`
	CollectedClues    []string             `json:"collected_clues"`
	UnlockedLocations []string             `json:"unlocked_locations"`
	DomainUnlocked    bool                 `json:"domain_unlocked"`
	NPCStates         map[string]*NPCState `json:"npc_states"`
	ChaosPool         int                  `json:"chaos_pool"`
	LooseEnds         int                  `json:"loose_ends"`
	LocationOverloads map[string]int       `json:"location_overloads"` // 地点过载追踪
	AnomalyStatus     string               `json:"anomaly_status"`
	MissionOutcome    string               `json:"mission_outcome"`
}

// NPCState NPC状态
type NPCState struct {
	ID              string                 `json:"id"`
	CurrentState    string                 `json:"current_state"`
	AnomalyAffected bool                   `json:"anomaly_affected"`
	Relationship    int                    `json:"relationship"`
	CustomData      map[string]interface{} `json:"custom_data"`
}
