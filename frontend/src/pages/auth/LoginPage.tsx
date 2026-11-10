import { useState } from 'react';
import { Form, Input, Button, Card, Typography, Message } from '@arco-design/web-react';
import { useNavigate } from 'react-router-dom';
import client from '../../api/client';

const { Title, Text } = Typography;

export default function LoginPage() {
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();
  const navigate = useNavigate();

  const handleLogin = async () => {
    setLoading(true);
    try {
      const values = await form.validate();
      const res = await client.post('/auth/login', values);
      localStorage.setItem('token', res.data.token);
      navigate('/dashboard', { replace: true });
    } catch (e: unknown) {
      if (e && typeof e === 'object' && 'response' in e) {
        const axErr = e as { response?: { data?: { message?: string } } };
        Message.error(axErr.response?.data?.message || '用户名或密码错误');
      } else if (!(e instanceof Error && e.message.includes('validate'))) {
        Message.error('登录失败，请检查服务器状态');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ minHeight: '100vh', background: '#F2F3F5', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
      <Card style={{ width: 400 }}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <Title heading={4} style={{ margin: 0 }}>SmartCDN</Title>
          <Text type="secondary">管理控制台</Text>
        </div>
        <Form form={form} layout="vertical" onSubmit={handleLogin}>
          <Form.Item label="用户名" field="username" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input placeholder="admin" />
          </Form.Item>
          <Form.Item label="密码" field="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password placeholder="密码" />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={loading} long style={{ marginTop: 8 }}>
            登录
          </Button>
        </Form>
      </Card>
    </div>
  );
}
