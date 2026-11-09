import React, { useState } from 'react';
import { Steps, Card, Button, Form, Input, Space, Typography, Alert, Spin } from '@arco-design/web-react';
import axios from 'axios';

const { Step } = Steps;
const { Title, Text } = Typography;

interface CheckItem {
  name: string;
  status: 'ok' | 'fail';
  message?: string;
}

interface SetupCheckResult {
  mysql: CheckItem;
  redis: CheckItem;
  go_version: CheckItem;
  ports: CheckItem;
  clock: CheckItem;
  permissions: CheckItem;
}

const STEP_LABELS = [
  '欢迎',
  '环境检测',
  '数据库配置',
  'Redis 配置',
  '数据迁移',
  '管理员账号',
  '加密密钥',
  '访问地址',
  '完成',
];

export default function SetupWizard() {
  const [current, setCurrent] = useState(0);
  const [loading, setLoading] = useState(false);
  const [checkResult, setCheckResult] = useState<SetupCheckResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [form] = Form.useForm();

  const runEnvCheck = async () => {
    setLoading(true);
    setError(null);
    try {
      const { data } = await axios.get<SetupCheckResult>('/api/v1/setup/check');
      setCheckResult(data);
      const hasFailure = Object.values(data).some((item) => item.status === 'fail');
      if (!hasFailure) {
        setCurrent(2);
      }
    } catch (e: unknown) {
      setError('环境检测失败，请检查服务器日志');
    } finally {
      setLoading(false);
    }
  };

  const handleComplete = async () => {
    setLoading(true);
    setError(null);
    try {
      const values = await form.validate();
      await axios.post('/api/v1/setup/complete', values);
      setCurrent(8);
    } catch (e: unknown) {
      if (axios.isAxiosError(e)) {
        setError(e.response?.data?.message || '安装失败');
      } else {
        setError('安装失败，请检查服务器日志');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ minHeight: '100vh', background: '#F2F3F5', display: 'flex', flexDirection: 'column', alignItems: 'center', paddingTop: 48 }}>
      <div style={{ width: '100%', maxWidth: 900, padding: '0 24px' }}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <Title heading={3} style={{ margin: 0 }}>SmartCDN 安装向导</Title>
          <Text type="secondary">完成以下步骤以初始化您的 SmartCDN 实例</Text>
        </div>

        <Steps current={current} style={{ marginBottom: 32 }} size="small">
          {STEP_LABELS.map((label, i) => (
            <Step key={i} title={label} />
          ))}
        </Steps>

        {error && (
          <Alert type="error" content={error} style={{ marginBottom: 16 }} closable onClose={() => setError(null)} />
        )}

        <Card>
          {/* Step 0: Welcome */}
          {current === 0 && (
            <div style={{ textAlign: 'center', padding: '32px 0' }}>
              <Title heading={4}>欢迎使用 SmartCDN</Title>
              <Text>SmartCDN 是可私有化部署的 DNS/CDN/SSL 综合管理平台。</Text>
              <Text>本向导将引导您完成首次安装配置，预计需要 5-10 分钟。</Text>
              <div style={{ marginTop: 24 }}>
                <Button type="primary" size="large" onClick={() => { setCurrent(1); runEnvCheck(); }}>
                  开始安装
                </Button>
              </div>
            </div>
          )}

          {/* Step 1: Environment Check */}
          {current === 1 && (
            <div>
              <Title heading={5}>环境检测</Title>
              {loading && <Spin tip="正在检测环境..." />}
              {checkResult && !loading && (
                <div>
                  {Object.entries(checkResult).map(([key, item]) => (
                    <div key={key} style={{ display: 'flex', alignItems: 'center', marginBottom: 8 }}>
                      <span style={{ color: item.status === 'ok' ? '#00B42A' : '#F53F3F', marginRight: 8, fontWeight: 600 }}>
                        {item.status === 'ok' ? '✓' : '✗'}
                      </span>
                      <Text>
                        {item.name}
                        {item.message ? ` — ${item.message}` : ''}
                      </Text>
                    </div>
                  ))}
                  {Object.values(checkResult).some((i) => i.status === 'fail') && (
                    <Alert type="warning" content="存在检测失败项，请修复后重试" style={{ marginTop: 16 }} />
                  )}
                </div>
              )}
              <Space style={{ marginTop: 16 }}>
                <Button onClick={runEnvCheck} loading={loading}>重新检测</Button>
                {checkResult && !Object.values(checkResult).some((i) => i.status === 'fail') && (
                  <Button type="primary" onClick={() => setCurrent(2)}>下一步</Button>
                )}
              </Space>
            </div>
          )}

          {/* Steps 2-7: Configuration Form */}
          {current >= 2 && current <= 7 && (
            <Form form={form} layout="vertical" autoComplete="off">
              {current === 2 && (
                <>
                  <Title heading={5}>数据库配置</Title>
                  <Form.Item label="主机" field="db_host" rules={[{ required: true }]}><Input placeholder="127.0.0.1" /></Form.Item>
                  <Form.Item label="端口" field="db_port" initialValue={3306} rules={[{ required: true }]}><Input type="number" /></Form.Item>
                  <Form.Item label="数据库名" field="db_name" rules={[{ required: true }]}><Input placeholder="smartcdn" /></Form.Item>
                  <Form.Item label="用户名" field="db_user" rules={[{ required: true }]}><Input /></Form.Item>
                  <Form.Item label="密码" field="db_password"><Input.Password /></Form.Item>
                </>
              )}
              {current === 3 && (
                <>
                  <Title heading={5}>Redis 配置</Title>
                  <Form.Item label="Redis 地址" field="redis_addr" rules={[{ required: true }]} initialValue="127.0.0.1:6379">
                    <Input />
                  </Form.Item>
                </>
              )}
              {current === 5 && (
                <>
                  <Title heading={5}>管理员账号</Title>
                  <Form.Item label="用户名" field="admin_username" rules={[{ required: true }]}><Input /></Form.Item>
                  <Form.Item
                    label="密码"
                    field="admin_password"
                    rules={[
                      { required: true },
                      { minLength: 8, message: '密码至少 8 位' },
                      {
                        validator: (v: string | undefined, cb: (err?: React.ReactNode) => void) => {
                          if (!v || !/[A-Z]/.test(v) || !/[0-9]/.test(v)) {
                            cb('密码须包含大写字母和数字');
                          } else {
                            cb();
                          }
                        },
                      },
                    ]}
                  >
                    <Input.Password />
                  </Form.Item>
                </>
              )}
              {current === 7 && (
                <>
                  <Title heading={5}>访问地址</Title>
                  <Form.Item label="站点 URL" field="site_url" rules={[{ required: true }, { type: 'url' }]} initialValue="http://localhost:8080">
                    <Input />
                  </Form.Item>
                </>
              )}

              <Space>
                {current > 2 && <Button onClick={() => setCurrent((c) => c - 1)}>上一步</Button>}
                {current < 7 && (
                  <Button type="primary" onClick={() => setCurrent((c) => c + 1)}>下一步</Button>
                )}
                {current === 7 && (
                  <Button type="primary" loading={loading} onClick={handleComplete}>开始安装</Button>
                )}
              </Space>
            </Form>
          )}

          {/* Step 8: Complete */}
          {current === 8 && (
            <div style={{ textAlign: 'center', padding: '32px 0' }}>
              <div style={{ fontSize: 48, color: '#00B42A', marginBottom: 16 }}>✓</div>
              <Title heading={4}>安装完成</Title>
              <Text>SmartCDN 已成功安装，您可以开始使用了。</Text>
              <div style={{ marginTop: 24 }}>
                <Button type="primary" size="large" href="/login">前往登录</Button>
              </div>
            </div>
          )}
        </Card>
      </div>
    </div>
  );
}
