import { useState } from 'react';
import {
  Button, Card, Table, Tag, Typography, Popconfirm, Message, Space, Modal, Form, Input, Select, InputNumber, Switch,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh, IconEdit } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface CDNSite {
  id: number; domain_name: string; provider_type: string;
  status: string; cname: string; origin_addr: string; origin_port: number; protocol: string;
  is_active: boolean; created_at: string;
}
interface CDNAccount { id: number; name: string; provider_type: string; }

const statusColor: Record<string, string> = {
  active: 'green', pending: 'orange', failed: 'red', inactive: 'gray',
};

export default function CDNSitesPage() {
  const qc = useQueryClient();
  const [createVisible, setCreateVisible] = useState(false);
  const [editVisible, setEditVisible] = useState(false);
  const [editingRow, setEditingRow] = useState<CDNSite | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();

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
      setCreateVisible(false);
      createForm.resetFields();
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '添加失败');
    },
  });

  const updateMut = useMutation({
    mutationFn: ({ id, ...body }: { id: number } & Record<string, unknown>) =>
      client.put(`/cdn/domains/${id}`, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['cdn-sites'] });
      Message.success('已更新');
      setEditVisible(false);
      editForm.resetFields();
      setEditingRow(null);
    },
    onError: (e: unknown) => {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error;
      Message.error(msg || '更新失败');
    },
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => client.delete(`/cdn/domains/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['cdn-sites'] }); Message.success('已删除'); },
    onError: () => Message.error('删除失败'),
  });

  const accounts: CDNAccount[] = accountsData?.items ?? [];

  const handleEdit = (row: CDNSite) => {
    setEditingRow(row);
    editForm.setFieldsValue({
      origin_addr: row.origin_addr,
      origin_port: row.origin_port,
      protocol: row.protocol,
      is_active: row.is_active,
    });
    setEditVisible(true);
  };

  const handleUpdate = async () => {
    if (!editingRow) return;
    const vals = await editForm.validate();
    updateMut.mutate({ id: editingRow.id, ...vals });
  };

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
        <Space>
          <Button size="mini" type="text" icon={<IconEdit />} onClick={() => handleEdit(row)}>编辑</Button>
          <Popconfirm title="确认删除此 CDN 站点？" onOk={() => deleteMut.mutate(row.id)}>
            <Button size="mini" type="text" status="danger">删除</Button>
          </Popconfirm>
        </Space>
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
            <Button type="primary" icon={<IconPlus />} onClick={() => setCreateVisible(true)}>添加站点</Button>
          </Space>
        }
      >
        <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
      </Card>

      {/* 新建弹窗 */}
      <Modal
        title="添加 CDN 站点"
        visible={createVisible}
        onOk={() => createForm.validate().then(vals => createMut.mutate(vals))}
        confirmLoading={createMut.isPending}
        onCancel={() => { setCreateVisible(false); createForm.resetFields(); }}
        unmountOnExit
      >
        <Form form={createForm} layout="vertical">
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
              <Select.Option value="follow">跟随请求</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>

      {/* 编辑弹窗 */}
      <Modal
        title="编辑 CDN 站点"
        visible={editVisible}
        onOk={handleUpdate}
        confirmLoading={updateMut.isPending}
        onCancel={() => { setEditVisible(false); editForm.resetFields(); setEditingRow(null); }}
        unmountOnExit
      >
        <Form form={editForm} layout="vertical">
          <Form.Item label="源站地址" field="origin_addr">
            <Input placeholder="1.2.3.4 或 origin.example.com" />
          </Form.Item>
          <Form.Item label="源站端口" field="origin_port">
            <InputNumber min={1} max={65535} />
          </Form.Item>
          <Form.Item label="协议" field="protocol">
            <Select>
              <Select.Option value="http">HTTP</Select.Option>
              <Select.Option value="https">HTTPS</Select.Option>
              <Select.Option value="follow">跟随请求</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="启用" field="is_active" triggerPropName="checked">
            <Switch checkedText="启用" uncheckedText="禁用" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
