package service

import (
	"testing"

	"github.com/trpg-solo-engine/backend/internal/domain"
)

// 测试心智控制检测
func TestIsMindControl(t *testing.T) {
	diceService := domain.NewDiceService()
	chaosService := NewChaosService()
	requestService := NewRequestService(diceService, chaosService)

	testCases := []struct {
		effect   string
		expected bool
	}{
		{"控制他的思想", true},
		{"操纵她的意志", true},
		{"强迫他们服从", true},
		{"命令他做某事", true},
		{"洗脑目标", true},
		{"催眠对方", true},
		{"支配他的心智", true},
		{"迫使她改变想法", true},
		{"让他想要帮助我", true},
		{"改变他们的思想", true},
		{"门突然打开了", false},
		{"我找到了一把钥匙", false},
	}

	for _, tc := range testCases {
		t.Run(tc.effect, func(t *testing.T) {
			result := requestService.IsMindControl(tc.effect)
			if result != tc.expected {
				t.Errorf("IsMindControl(%q) = %v, expected %v", tc.effect, result, tc.expected)
			}
		})
	}
}

// 测试请求验证
func TestValidateRequest(t *testing.T) {
	diceService := domain.NewDiceService()
	chaosService := NewChaosService()
	requestService := NewRequestService(diceService, chaosService)

	t.Run("有效的请求", func(t *testing.T) {
		req := &RealityChangeRequest{
			Effect:      "门突然打开了",
			CausalChain: "因为我之前在门上做了手脚，所以现在门锁失效了",
			Quality:     "专注",
			LocationID:  "test_location",
		}

		err := requestService.ValidateRequest(req)
		if err != nil {
			t.Errorf("ValidateRequest() error = %v, expected nil", err)
		}
	})

	t.Run("心智控制应该被拒绝", func(t *testing.T) {
		req := &RealityChangeRequest{
			Effect:      "控制他的思想",
			CausalChain: "这是一个足够长的因果链描述，用于测试心智控制检测",
			Quality:     "专注",
			LocationID:  "test_location",
		}

		err := requestService.ValidateRequest(req)
		if err == nil {
			t.Error("ValidateRequest() should return error for mind control, but got nil")
		}
	})
}
