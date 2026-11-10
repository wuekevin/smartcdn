import { Button, Card, Table, Tag, Typography, Popconfirm, Message } from '@arco-design/web-react';
import { IconRefresh } from '@arco-design/web-react/icon';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import client from '../../api/client';

interface CDNSite {
  id: number; domain_name: string; provider_type: string;
  status: string; cname: string; created_at: string;
}

const statusColor: Record<string, string> = {
  active: 'green', pending: 'orange', failed: 'red', inactive: 'gray',
};

export default function CDNSitesPage() {
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['cdn-sites'],
    queryFn: () => client.get('/cdn/sites').then(r => r.data),
  });

  const deleteMut = useMutation({
    mutationFn: (id: number) => client.delete(`/cdn/sites/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['cdn-sites'] }); Message.success('已删除'); },
    onError: () => Message.error('删除失败'),
  });

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
    <Card
      title={<Typography.Title heading={6} style={{ margin: 0 }}>CDN 站点</Typography.Title>}
      extra={<Button icon={<IconRefresh />} onClick={() => qc.invalidateQueries({ queryKey: ['cdn-sites'] })} />}
    >
      <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
    </Card>
  );
}
