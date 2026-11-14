import { useState } from 'react';
import {
  Button, Card, Table, Tag, Typography, Popconfirm, Message, Space, Modal, Form, Input, Select, InputNumber,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface CDNSite {
  id: number; domain_name: string; provider_type: string;
  status: string; cname: string; created_at: string;
}
interface CDNAccount { id: number; name: string; provider_type: string; }

const statusColor: Record<string, string> = {
  active: 'green', pending: 'orange', failed: 'red', inactive: 'gray',
};

export default function CDNSitesPage() {
  const qc = useQueryClient();
  const [visible, setVisible] = useState(false);
  const [form] = Form.useForm();

  const { data, isLoading } = useQuery({
    queryKey: ['cdn-sites'],
    queryFn: () => client.get('/cdn/domains').then(r => r.data),
  });

  const { data: accountsData } = useQuery({
    queryKey: ['cdn-accounts'],
    queryFn: () => client.get('/cdn/accounts').then(r => r.data),
  });

  const createMut = useMutation({
    mutationFn: (body: unknown) => client.post('/cdn/domains', body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['cdn-sites'] });
      Message.success('CDN 站点已添加');
      setVisible(false);
      form.resetFields();
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '添加失败');
    },
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => client.delete(`/cdn/domains/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['cdn-sites'] }); Message.success('已删除'); },
    onError: () => Message.error('删除失败'),
  });

  const accounts: CDNAccount[] = accountsData?.items ?? [];

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '域名', dataIndex: 'domain_name' },
    { title: '服务商', dataIndex: 'provider_type' },
    {
      title: '状态', dataIndex: 'status',
      render: (v: string) => <Tag color={statusColor[v] ?? 'gray'}>{v}</Tag>,
    },
    { title: 'CNAME', dataIndex: 'cname' },
    { title: '创建时间', dataIndex: 'created_at', width: 180 },
    {
      title: '操作', render: (_: unknown, row: CDNSite) => (
        <Popconfirm title="确认删除此 CDN 站点？" onOk={() => deleteMut.mutate(row.id)}>
          <Button size="mini" type="text" status="danger">删除</Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <>
      <Card
        title={<Typography.Title heading={6} style={{ margin: 0 }}>CDN 站点</Typography.Title>}
        extra={
          <Space>
            <Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['cdn-sites'] })} />
            <Button type="primary" icon={<IconPlus />} onClick={() => setVisible(true)}>添加站点</Button>
          </Space>
        }
      >
        <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
      </Card>

      <Modal
        title="添加 CDN 站点"
        visible={visible}
        onOk={() => form.validate().then(vals => createMut.mutate(vals))}
        confirmLoading={createMut.isPending}
        onCancel={() => { setVisible(false); form.resetFields(); }}
        unmountOnExit
      >
        <Form form={form} layout="vertical">
          <Form.Item label="CDN 账户" field="account_id" rules={[{ required: true, message: '请选择 CDN 账户' }]}>
            <Select placeholder="选择 CDN 账户">
              {accounts.map(a => (
                <Select.Option key={a.id} value={a.id}>{a.name} ({a.provider_type})</Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item label="域名" field="domain_name" rules={[{ required: true, message: '请输入域名' }]}>
            <Input placeholder="cdn.example.com" />
          </Form.Item>
          <Form.Item label="源站地址" field="origin_addr">
            <Input placeholder="1.2.3.4 或 origin.example.com" />
          </Form.Item>
          <Form.Item label="源站端口" field="origin_port" initialValue={80}>
            <InputNumber min={1} max={65535} />
          </Form.Item>
          <Form.Item label="协议" field="protocol" initialValue="http">
            <Select>
              <Select.Option value="http">HTTP</Select.Option>
              <Select.Option value="https">HTTPS</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
