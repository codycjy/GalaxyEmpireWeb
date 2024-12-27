package models

import "time"

// BalanceLogType 余额变动类型
type BalanceLogType string

const (
	BalanceLogTypeDeposit    BalanceLogType = "deposit"    // 充值
	BalanceLogTypeWithdraw   BalanceLogType = "withdraw"   // 提现
	BalanceLogTypeSpent      BalanceLogType = "spent"      // 消费
	BalanceLogTypeRefund     BalanceLogType = "refund"     // 退款
	BalanceLogTypeAdjustment BalanceLogType = "adjustment" // 调整
)

// BalanceLog 余额变动记录
type BalanceLog struct {
	ID          uint           `gorm:"primarykey"`
	UserID      uint           `gorm:"index:idx_user_created,priority:1"` // 用户ID
	Amount      int64          `gorm:"not null"`                          // 变动金额（正数表示增加，负数表示减少）
	Balance     int64          `gorm:"not null"`                          // 变动后的余额
	Type        BalanceLogType `gorm:"type:string;not null;index"`        // 变动类型
	Reference   string         `gorm:"type:string;index"`                 // 关联ID（如支付ID、订单ID等）
	Description string         `gorm:"type:string"`                       // 变动说明
	CreatedAt   time.Time      `gorm:"index:idx_user_created,priority:2"` // 创建时间
	UpdatedAt   time.Time
}
