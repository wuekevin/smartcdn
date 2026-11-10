import { Card, Table, Typography } from '@arco-design/web-react';
import { useQuery } from '@tanstack/react-query';
import client from '../../api/client';

export default function AuditPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['audit-logs'],
    queryFn: () => client.get('/audit-logs', { params: { limit: 50 } }).then(r => r.data),
    refetchInterval: 30_000,
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    { title: '操作者 ID', dataIndex: 'operator_id', width: 100 },
    { title: '操作', dataIndex: 'action' },
    { title: '资源类型', dataIndex: 'resource_type' },
    { title: '客户端 IP', dataIndex: 'operator_ip', width: 140 },
    { title: '时间', dataIndex: 'created_at', width: 180 },
  ];

  return (
    <Card title={<Typography.Title heading={6} style={{ margin: 0 }}>审计日志</Typography.Title>}>
      <Table loading={isLoading} columns={columns} data={data?.items ?? []} rowKey="id" pagination={false} />
    </Card>
  );
}
