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

// Fields that must pass validation before leaving each step (0-indexed).
const STEP_REQUIRED_FIELDS: Record<number, string[]> = {
  2: ['db_host', 'db_port', 'db_name', 'db_user'],
  3: ['redis_addr'],
  5: ['admin_username', 'admin_password'],
  7: ['port', 'site_url'],
};

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

function genHexKey(): string {
  const bytes = new Uint8Array(32);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, b => b.toString(16).padStart(2, '0')).join('');
}

export default function SetupWizard() {
  const [current, setCurrent] = useState(0);
  const [loading, setLoading] = useState(false);
  const [checkResult, setCheckResult] = useState<SetupCheckResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [form] = Form.useForm();
  // Generated once; stable across re-renders
  const [encryptKey] = useState(genHexKey);

  const runEnvCheck = async () => {
    setLoading(true);
    setError(null);
    try {
      const { data } = await axios.get<SetupCheckResult>('/api/v1/setup/check');
      setCheckResult(data);
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
      const raw = await form.validate();
      const values = {
        ...raw,
        db_port: Number(raw.db_port),
        redis_db: Number(raw.redis_db),
      };
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

  const copyKey = () => {
    navigator.clipboard?.writeText(encryptKey).catch(() => {});
  };

  return (
    <div style={{ minHeight: '100vh', background: '#F2F3F5', display: 'flex', flexDirection: 'column', alignItems: 'center', paddingTop: 48 }}>
      <div style={{ width: '100%', maxWidth: 900, padding: '0 24px' }}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <Title heading={3} style={{ margin: 0 }}>SmartCDN 安装向导</Title>
          <Text type="secondary">完成以下步骤以初始化您的 SmartCDN 实例</Text>
        </div>

        {/* Arco Design Steps is 1-indexed: pass current + 1 so the highlighted
            step matches the rendered content (which uses 0-indexed state). */}
        <Steps current={current + 1} style={{ marginBottom: 32 }} size="small">
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

          {/* Steps 2–7: Configuration Form
              All step sections stay mounted once the form appears so that
              Arco Design Form does not lose field values on step change.
              Hidden steps are rendered with display:none. */}
          {current >= 2 && current <= 7 && (
            <Form form={form} layout="vertical" autoComplete="off">

              {/* Step 2: 数据库配置 */}
              <div style={{ display: current === 2 ? undefined : 'none' }}>
                <Title heading={5}>数据库配置</Title>
                <Form.Item label="主机" field="db_host" rules={[{ required: true }]}>
                  <Input placeholder="127.0.0.1" />
                </Form.Item>
                <Form.Item label="端口" field="db_port" initialValue={3306} rules={[{ required: true }]}>
                  <Input type="number" />
                </Form.Item>
                <Form.Item label="数据库名" field="db_name" rules={[{ required: true }]}>
                  <Input placeholder="smartcdn" />
                </Form.Item>
                <Form.Item label="用户名" field="db_user" rules={[{ required: true }]}>
                  <Input />
                </Form.Item>
                <Form.Item label="密码" field="db_password">
                  <Input.Password />
                </Form.Item>
              </div>

              {/* Step 3: Redis 配置 */}
              <div style={{ display: current === 3 ? undefined : 'none' }}>
                <Title heading={5}>Redis 配置</Title>
                <Form.Item label="Redis 地址" field="redis_addr" rules={[{ required: true }]} initialValue="127.0.0.1:6379">
                  <Input placeholder="127.0.0.1:6379" />
                </Form.Item>
                <Form.Item label="Redis 密码" field="redis_password">
                  <Input.Password placeholder="无密码留空" />
                </Form.Item>
                <Form.Item label="Redis DB 编号" field="redis_db" initialValue={0}>
                  <Input type="number" placeholder="0" style={{ width: 120 }} />
                </Form.Item>
              </div>

              {/* Step 4: 数据迁移（无表单字段） */}
              {current === 4 && (
                <div>
                  <Title heading={5}>数据迁移</Title>
                  <Alert
                    type="info"
                    style={{ marginBottom: 16 }}
                    content="数据库表结构将在安装完成时自动创建，无需任何手动操作。"
                  />
                  <Text type="secondary">点击「下一步」继续完成安装配置。</Text>
                </div>
              )}

              {/* Step 5: 管理员账号 */}
              <div style={{ display: current === 5 ? undefined : 'none' }}>
                <Title heading={5}>管理员账号</Title>
                <Form.Item label="用户名" field="admin_username" rules={[{ required: true }]}>
                  <Input />
                </Form.Item>
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
              </div>

              {/* Step 6: 加密密钥 */}
              <div style={{ display: current === 6 ? undefined : 'none' }}>
                <Title heading={5}>加密密钥</Title>
                <Alert
                  type="warning"
                  style={{ marginBottom: 16 }}
                  content="系统使用 AES-256-GCM 加密所有 API 凭证和 SSL 私钥。安装完成后请将以下密钥设置为环境变量 SMARTCDN_ENCRYPT_KEY，丢失后加密数据将无法恢复。"
                />
                <Form.Item
                  label="加密密钥（SMARTCDN_ENCRYPT_KEY）"
                  field="encrypt_key"
                  initialValue={encryptKey}
                  extra="已为您自动生成 32 字节随机密钥，请复制保存到安全位置。"
                >
                  <Input
                    readOnly
                    suffix={
                      <Button size="mini" type="text" onClick={copyKey}>
                        复制
                      </Button>
                    }
                  />
                </Form.Item>
              </div>

              {/* Step 7: 访问地址 */}
              <div style={{ display: current === 7 ? undefined : 'none' }}>
                <Title heading={5}>访问地址</Title>
                <Form.Item
                  label="监听端口"
                  field="port"
                  initialValue="8080"
                  rules={[
                    { required: true },
                    {
                      validator: (v: string | undefined, cb: (err?: React.ReactNode) => void) => {
                        const n = Number(v);
                        if (!v || isNaN(n) || n < 1 || n > 65535) {
                          cb('端口范围 1–65535');
                        } else {
                          cb();
                        }
                      },
                    },
                  ]}
                  extra="SmartCDN HTTP 服务监听的端口号"
                >
                  <Input type="number" style={{ width: 160 }} />
                </Form.Item>
                <Form.Item
                  label="站点公网 URL"
                  field="site_url"
                  rules={[{ required: true }, { type: 'url' }]}
                  initialValue="http://localhost:8080"
                  extra="对外访问的完整地址，用于生成回调链接等"
                >
                  <Input placeholder="https://cdn.example.com" />
                </Form.Item>
              </div>

              <Space style={{ marginTop: 16 }}>
                {current > 1 && (
                  <Button onClick={() => setCurrent((c) => c - 1)}>上一步</Button>
                )}
                {current < 7 && (
                  <Button
                    type="primary"
                    onClick={async () => {
                      const fields = STEP_REQUIRED_FIELDS[current];
                      if (fields) {
                        try {
                          await form.validate(fields);
                        } catch {
                          return;
                        }
                      }
                      setCurrent((c) => c + 1);
                    }}
                  >
                    下一步
                  </Button>
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
              <Text>数据库已初始化，管理员账户已创建。</Text>
              <div style={{ marginTop: 16, textAlign: 'left', background: '#f5f5f5', borderRadius: 6, padding: '16px 24px', maxWidth: 480, margin: '16px auto 0' }}>
                <Text bold>下一步操作：</Text>
                <ol style={{ marginTop: 8, paddingLeft: 20 }}>
                  <li>配置已保存至服务器 <code>config/config.yaml</code>，请妥善保管</li>
                  <li>重启 SmartCDN 服务（<code>systemctl restart smartcdn</code>）</li>
                  <li>访问 <code>/login</code> 使用管理员账户登录</li>
                </ol>
              </div>
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
