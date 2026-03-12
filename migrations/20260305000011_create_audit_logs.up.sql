-- Create audit_logs table for security and compliance
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    username VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    resource VARCHAR(50) NOT NULL,
    resource_id UUID,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    before_data JSONB,
    after_data JSONB,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for efficient querying
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX idx_audit_logs_resource_id ON audit_logs(resource_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_user_action ON audit_logs(user_id, action);
CREATE INDEX idx_audit_logs_resource_resource_id ON audit_logs(resource, resource_id);

-- Create a composite index for common queries
CREATE INDEX idx_audit_logs_composite ON audit_logs(user_id, resource, action, created_at DESC);

COMMENT ON TABLE audit_logs IS '审计日志表，记录所有重要操作';
COMMENT ON COLUMN audit_logs.user_id IS '操作用户ID';
COMMENT ON COLUMN audit_logs.username IS '操作用户名（冗余存储，防止用户删除后无法追溯）';
COMMENT ON COLUMN audit_logs.action IS '操作类型：create, update, delete, login, logout, view, export';
COMMENT ON COLUMN audit_logs.resource IS '资源类型：user, game, release, product, order, comment, achievement, coupon';
COMMENT ON COLUMN audit_logs.resource_id IS '资源ID';
COMMENT ON COLUMN audit_logs.ip_address IS '操作IP地址';
COMMENT ON COLUMN audit_logs.user_agent IS '用户代理字符串';
COMMENT ON COLUMN audit_logs.before_data IS '操作前数据（JSON格式）';
COMMENT ON COLUMN audit_logs.after_data IS '操作后数据（JSON格式）';
COMMENT ON COLUMN audit_logs.error_message IS '错误信息（如果操作失败）';
