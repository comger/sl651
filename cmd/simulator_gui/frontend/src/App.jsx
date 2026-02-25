import React, { useState, useEffect, useRef } from 'react';
import { Layout, Menu, Card, Form, Input, InputNumber, Switch, Button, Table, Tabs, Tag, Space, Typography, List, Divider, message, Select, Modal } from 'antd';
import {
    PlayCircleOutlined,
    StopOutlined,
    SettingOutlined,
    ThunderboltOutlined,
    DesktopOutlined,
    HistoryOutlined,
    SaveOutlined,
    PlusOutlined,
    DeleteOutlined,
    ReloadOutlined
} from '@ant-design/icons';
import { GetConfig, SaveConfig, StartSimulation, StopSimulation, AddDevice, DeleteDevice } from "../wailsjs/go/main/App";
import { EventsOn } from "../wailsjs/runtime/runtime";
import './App.css';

const { Header, Content, Sider } = Layout;
const { Title, Text } = Typography;

const STANDARD_FACTORS = [
    { name: "水位", tag: 0x39, decimals: 3, algo: "sin", base: 10, min: 9, max: 11 },
    { name: "雨量", tag: 0x26, decimals: 1, algo: "step", base: 0, step: 0.5 },
    { name: "电压", tag: 0x38, decimals: 2, algo: "random", base: 12.5, min: 12, max: 13 },
    { name: "流量", tag: 0x37, decimals: 3, algo: "constant", base: 100 },
    { name: "风速", tag: 0x20, decimals: 1, algo: "random", base: 2.5, min: 0, max: 10 },
    { name: "气温", tag: 0x22, decimals: 1, algo: "sin", base: 25, min: 20, max: 30 },
];

function App() {
    const [config, setConfig] = useState(null);
    const [selectedDeviceIndex, setSelectedDeviceIndex] = useState(0);
    const [logs, setLogs] = useState([]);
    const [isRunning, setIsRunning] = useState(false);
    const [addDeviceModalVisible, setAddDeviceModalVisible] = useState(false);
    const [addChannelModalVisible, setAddChannelModalVisible] = useState(false);
    const [form] = Form.useForm();
    const [addDeviceForm] = Form.useForm();
    const [channelForm] = Form.useForm(); // Single form for Add/Edit
    const [editingTag, setEditingTag] = useState(null); // Track if editing specific tag

    const refreshConfig = () => {
        GetConfig().then(cfg => {
            setConfig(cfg);
            if (cfg && cfg.devices && cfg.devices.length > 0) {
                const idx = selectedDeviceIndex < cfg.devices.length ? selectedDeviceIndex : 0;
                setSelectedDeviceIndex(idx);
                form.setFieldsValue(cfg.devices[idx]);
            }
        });
    };

    useEffect(() => {
        const init = async () => {
            try {
                // Check if running in Wails context
                if (!window['go']) {
                    console.warn("Wails runtime not found. Using mock config.");
                    // Mock config for browser testing
                    setConfig({
                        server_addr: "127.0.0.1:8085 (Browser Mode)",
                        devices: [
                            {
                                id: "MOCK_DEVICE",
                                name: "浏览器测试站",
                                interval_seconds: 60,
                                channels: []
                            }
                        ]
                    });
                    return;
                }

                const cfg = await GetConfig();
                setConfig(cfg);
                if (cfg && cfg.devices && cfg.devices.length > 0) {
                    const idx = selectedDeviceIndex < cfg.devices.length ? selectedDeviceIndex : 0;
                    setSelectedDeviceIndex(idx);
                    form.setFieldsValue(cfg.devices[idx]);
                }

                EventsOn("simulator:log", (log) => {
                    setLogs(prev => [log, ...prev].slice(0, 100));
                });
            } catch (err) {
                console.error("Initialization error:", err);
                message.error("初始化失败: " + err.message);
            }
        };
        init();
    }, []);

    const handleDeviceSelect = (index) => {
        setSelectedDeviceIndex(index);
        form.setFieldsValue(config.devices[index]);
    };

    const handleAddDevice = async (values) => {
        // Mock Mode: Add to local state
        if (!window['go']) {
            const newDevice = {
                id: "MOCK_" + Date.now(),
                name: values.name,
                interval_seconds: 60,
                channels: []
            };
            const newConfig = { ...config, devices: [...config.devices, newDevice] };
            setConfig(newConfig);
            setAddDeviceModalVisible(false);
            addDeviceForm.resetFields();
            message.success("站点已添加 (浏览器模式)");
            return;
        }

        await AddDevice(values.name);
        setAddDeviceModalVisible(false);
        addDeviceForm.resetFields();
        message.success("站点已添加");
        refreshConfig();
    };

    const handleDeleteDevice = async (index) => {
        // Mock Mode: Remove from local state
        if (!window['go']) {
            const newConfig = { ...config };
            newConfig.devices = newConfig.devices.filter((_, i) => i !== index);
            setConfig(newConfig);
            // Reset selection if needed
            if (selectedDeviceIndex >= newConfig.devices.length) {
                setSelectedDeviceIndex(Math.max(0, newConfig.devices.length - 1));
            }
            message.info("站点已删除 (浏览器模式)");
            return;
        }

        await DeleteDevice(index);
        message.info("站点已删除");
        refreshConfig();
    };

    const handleSaveChannel = async (values) => {
        let newChannels = [...config.devices[selectedDeviceIndex].channels];

        // Remove existing if editing (or if tag changed)
        if (editingTag !== null) {
            newChannels = newChannels.filter(c => c.tag !== editingTag);
        } else {
            // If adding new, remove any existing with same tag to prevent duplicates
            newChannels = newChannels.filter(c => c.tag !== values.tag);
        }

        const newChannel = {
            tag: values.tag,
            name: values.name,
            decimals: values.decimals,
            algorithm: {
                type: values.algo_type,
                base: values.base,
                min: values.min || 0,
                max: values.max || 0,
                step: values.step || 0,
                scale: values.scale || 0
            }
        };
        newChannels.push(newChannel);

        const newConfig = { ...config };
        newConfig.devices[selectedDeviceIndex].channels = newChannels;

        // Mock Mode: Update local state only
        if (!window['go']) {
            setConfig(newConfig);
            message.success("因子配置已保存 (浏览器模式)");
            setEditingTag(null);
            channelForm.resetFields();
            return;
        }

        await SaveConfig(newConfig);
        setConfig(newConfig);
        message.success("因子配置已保存");
        setEditingTag(null);
        channelForm.resetFields();
    };

    const handleEditChannel = (channel) => {
        setEditingTag(channel.tag);
        channelForm.setFieldsValue({
            tag: channel.tag,
            name: channel.name,
            decimals: channel.decimals,
            algo_type: channel.algorithm.type,
            base: channel.algorithm.base,
            min: channel.algorithm.min,
            max: channel.algorithm.max,
            step: channel.algorithm.step,
            scale: channel.algorithm.scale
        });
    };

    const handleDeleteChannel = async (tag) => {
        const newConfig = { ...config };
        newConfig.devices[selectedDeviceIndex].channels = newConfig.devices[selectedDeviceIndex].channels.filter(c => c.tag !== tag);

        // Mock Mode
        if (!window['go']) {
            setConfig(newConfig);
            message.info("因子已移除 (浏览器模式)");
            if (editingTag === tag) {
                setEditingTag(null);
                channelForm.resetFields();
            }
            return;
        }

        await SaveConfig(newConfig);
        setConfig(newConfig);
        message.info("因子已移除");
        if (editingTag === tag) {
            setEditingTag(null);
            channelForm.resetFields();
        }
    };

    const handleSave = async (values) => {
        const newConfig = { ...config };
        newConfig.devices[selectedDeviceIndex] = {
            ...newConfig.devices[selectedDeviceIndex],
            ...values,
            channels: config.devices[selectedDeviceIndex].channels // Preserve channels for now
        };

        // Mock Mode
        if (!window['go']) {
            setConfig(newConfig);
            message.success("配置已保存 (浏览器模式)");
            return;
        }

        try {
            await SaveConfig(newConfig);
            setConfig(newConfig);
            message.success("配置已保存并落盘");
        } catch (err) {
            message.error("保存失败: " + err);
        }
    };

    const handleStart = () => {
        if (!window['go']) { message.success("模拟引擎已启动 (浏览器虚拟模式)"); setIsRunning(true); return; }
        StartSimulation();
        setIsRunning(true);
        message.success("模拟引擎已启动");
    };

    const handleStop = () => {
        if (!window['go']) { message.info("模拟引擎已停止 (浏览器虚拟模式)"); setIsRunning(false); return; }
        StopSimulation();
        setIsRunning(false);
        message.info("模拟引擎已停止并保存状态");
    };

    if (!config) return <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', background: '#001529', color: '#fff' }}>Loading Simulator...</div>;

    const currentDevice = (config.devices && config.devices[selectedDeviceIndex]) || null;

    return (
        <Layout style={{ height: '100vh' }}>
            <Header style={{ background: '#001529', padding: '0 20px', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Space>
                    <ThunderboltOutlined style={{ color: '#faad14', fontSize: '24px' }} />
                    <Title level={4} style={{ color: '#fff', margin: 0 }}>SL651 协协议模拟器 V2.1</Title>
                </Space>
                <Space>
                    <Text style={{ color: '#fff' }}>北向地址: {config.server_addr}</Text>
                    <Divider type="vertical" />
                    {!isRunning ? (
                        <Button type="primary" icon={<PlayCircleOutlined />} onClick={handleStart} success>启动模拟</Button>
                    ) : (
                        <Button danger icon={<StopOutlined />} onClick={handleStop}>停止模拟</Button>
                    )}
                </Space>
            </Header>

            <Layout>
                <Sider width={250} style={{ background: '#fff', borderRight: '1px solid #f0f0f0' }}>
                    <div style={{ padding: '16px', fontWeight: 'bold' }}>站点列表</div>
                    <Menu
                        mode="inline"
                        selectedKeys={[selectedDeviceIndex.toString()]}
                        onClick={({ key }) => handleDeviceSelect(parseInt(key))}
                        items={config.devices.map((d, i) => ({
                            key: i.toString(),
                            icon: <DesktopOutlined />,
                            label: (
                                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                    <span>{d.name || d.id}</span>
                                    <DeleteOutlined
                                        style={{ color: '#ff4d4f' }}
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            handleDeleteDevice(i);
                                        }}
                                    />
                                </div>
                            )
                        }))}
                    />
                    <div style={{ padding: '16px' }}>
                        <Button type="dashed" block icon={<PlusOutlined />} onClick={() => setAddDeviceModalVisible(true)}>添加站点</Button>
                    </div>
                </Sider>

                <Content style={{ display: 'flex', flexDirection: 'column' }}>
                    {!currentDevice ? (
                        <div style={{ flex: 1, display: 'flex', justifyContent: 'center', alignItems: 'center', color: '#999' }}>
                            <Space direction="vertical" align="center">
                                <DesktopOutlined style={{ fontSize: 48 }} />
                                <Text type="secondary">请先在左侧添加或选择一个站点</Text>
                            </Space>
                        </div>
                    ) : (
                        <div style={{ padding: '20px', flex: 1, overflowY: 'auto' }}>
                            <Tabs defaultActiveKey="1" items={[
                                {
                                    key: '1',
                                    label: '基础配置',
                                    children: (
                                        <Card bordered={false}>
                                            <Form form={form} layout="vertical" onFinish={handleSave}>
                                                <Space size="large" align="start">
                                                    <Form.Item label="观测站代号" name="id" rules={[{ required: true }]}>
                                                        <Input style={{ width: 200 }} />
                                                    </Form.Item>
                                                    <Form.Item label="站点名称" name="name">
                                                        <Input style={{ width: 200 }} />
                                                    </Form.Item>
                                                    <Form.Item label="查询密码" name="password">
                                                        <Input style={{ width: 120 }} maxLength={4} />
                                                    </Form.Item>
                                                </Space>

                                                <Space size="large">
                                                    <Form.Item label="上报间隔 (秒)" name="interval_seconds">
                                                        <InputNumber min={1} />
                                                    </Form.Item>
                                                    <Form.Item label="登入报" name="login_enabled" valuePropName="checked">
                                                        <Switch />
                                                    </Form.Item>
                                                    <Form.Item label="心跳报" name="heartbeat_enabled" valuePropName="checked">
                                                        <Switch />
                                                    </Form.Item>
                                                    <Form.Item label="小时报" name="hour_report_enabled" valuePropName="checked">
                                                        <Switch />
                                                    </Form.Item>
                                                </Space>

                                                <Form.Item>
                                                    <Button type="primary" htmlType="submit" icon={<SaveOutlined />}>应用并保存</Button>
                                                </Form.Item>
                                            </Form>
                                        </Card>
                                    )
                                },
                                {
                                    key: '2',
                                    label: '通道与因子',
                                    children: (
                                        <div style={{ display: 'flex', gap: '20px' }}>
                                            {/* Left Side: Table */}
                                            <div style={{ flex: 1 }}>
                                                <Table
                                                    dataSource={currentDevice.channels}
                                                    rowKey="tag"
                                                    pagination={false}
                                                    size="small"
                                                    onRow={(record) => ({
                                                        onClick: () => handleEditChannel(record),
                                                        style: { cursor: 'pointer', background: editingTag === record.tag ? '#e6f7ff' : 'transparent' }
                                                    })}
                                                    columns={[
                                                        { title: 'Tag', dataIndex: 'tag', render: t => <Text code>{t.toString(16).toUpperCase()}H</Text> },
                                                        { title: '名称', dataIndex: 'name' },
                                                        { title: '算法', dataIndex: ['algorithm', 'type'], render: t => <Tag color="blue">{t}</Tag> },
                                                        { title: '基值', dataIndex: ['algorithm', 'base'] },
                                                        { title: '操作', render: (_, r) => <Button type="link" danger icon={<DeleteOutlined />} onClick={(e) => { e.stopPropagation(); handleDeleteChannel(r.tag); }} size="small" /> }
                                                    ]}
                                                />
                                                <div style={{ marginTop: 10, textAlign: 'center' }}>
                                                    <Button type="dashed" onClick={() => { setEditingTag(null); channelForm.resetFields(); }}>+ 新增因子</Button>
                                                </div>
                                            </div>

                                            {/* Right Side: Edit Form */}
                                            <Card
                                                title={editingTag ? "编辑因子" : "新增因子"}
                                                size="small"
                                                style={{ width: 350, height: '100%' }}
                                                extra={
                                                    editingTag && <Button type="text" size="small" onClick={() => { setEditingTag(null); channelForm.resetFields(); }}>取消</Button>
                                                }
                                            >
                                                <Form form={channelForm} layout="vertical" onFinish={handleSaveChannel} initialValues={{ decimals: 3, algo_type: 'sin', base: 10 }}>
                                                    <Form.Item label="预设因子" style={{ marginBottom: 12 }}>
                                                        <Select
                                                            placeholder="快速选择标准因子..."
                                                            onChange={(idx) => {
                                                                const f = STANDARD_FACTORS[idx];
                                                                channelForm.setFieldsValue({
                                                                    name: f.name,
                                                                    tag: f.tag,
                                                                    decimals: f.decimals,
                                                                    algo_type: f.algo,
                                                                    base: f.base,
                                                                    min: f.min,
                                                                    max: f.max,
                                                                    step: f.step
                                                                });
                                                            }}
                                                        >
                                                            {STANDARD_FACTORS.map((f, i) => (
                                                                <Select.Option key={i} value={i}>{f.name} ({f.tag.toString(16).toUpperCase()}H)</Select.Option>
                                                            ))}
                                                        </Select>
                                                    </Form.Item>

                                                    <Space>
                                                        <Form.Item label="名称" name="name" rules={[{ required: true }]}>
                                                            <Input style={{ width: 120 }} />
                                                        </Form.Item>
                                                        <Form.Item label="Tag (Hex)" name="tag" rules={[{ required: true }]}>
                                                            <InputNumber min={0} max={255} formatter={v => v ? v.toString(16).toUpperCase() : ''} parser={v => parseInt(v, 16)} style={{ width: 80 }} />
                                                        </Form.Item>
                                                    </Space>

                                                    <Space>
                                                        <Form.Item label="算法" name="algo_type">
                                                            <Select style={{ width: 100 }}>
                                                                <Select.Option value="constant">固定</Select.Option>
                                                                <Select.Option value="random">随机</Select.Option>
                                                                <Select.Option value="sin">正弦</Select.Option>
                                                                <Select.Option value="step">步进</Select.Option>
                                                            </Select>
                                                        </Form.Item>
                                                        <Form.Item label="小数位" name="decimals">
                                                            <InputNumber min={0} max={4} style={{ width: 60 }} />
                                                        </Form.Item>
                                                    </Space>

                                                    <Space>
                                                        <Form.Item label="基值" name="base" rules={[{ required: true }]}>
                                                            <InputNumber style={{ width: 90 }} />
                                                        </Form.Item>
                                                        <Form.Item label="最小值" name="min">
                                                            <InputNumber style={{ width: 90 }} />
                                                        </Form.Item>
                                                    </Space>
                                                    <Space>
                                                        <Form.Item label="最大值" name="max">
                                                            <InputNumber style={{ width: 90 }} />
                                                        </Form.Item>
                                                        <Form.Item label="步进" name="step">
                                                            <InputNumber style={{ width: 90 }} step={0.1} />
                                                        </Form.Item>
                                                    </Space>

                                                    <Button type="primary" htmlType="submit" block icon={<SaveOutlined />}>保存因子配置</Button>
                                                </Form>
                                            </Card>
                                        </div>
                                    )
                                }
                            ]} />
                        </div>

                    )}
                    <div style={{ height: '300px', background: '#1e1e1e', borderTop: '4px solid #333', display: 'flex', flexDirection: 'column' }}>
                        <div style={{ padding: '8px 16px', background: '#252526', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                            <Space>
                                <HistoryOutlined style={{ color: '#ccc' }} />
                                <Text style={{ color: '#ccc', fontSize: '12px' }}>实时通信轨迹 (最后100条)</Text>
                            </Space>
                            <Button type="text" size="small" style={{ color: '#ccc' }} onClick={() => setLogs([])}>清空日志</Button>
                        </div>
                        <div style={{ flex: 1, overflowY: 'auto', padding: '10px' }} className="log-container">
                            <List
                                dataSource={[...logs].reverse()}
                                size="small"
                                renderItem={item => (
                                    <div style={{ marginBottom: '4px', fontFamily: 'monospace', fontSize: '12px' }}>
                                        <Text type="secondary">[{item.time}]</Text>
                                        <Tag color="cyan" style={{ marginLeft: 8 }}>{item.device_id}</Tag>
                                        <Tag color="orange" style={{ margin: '0 4px' }}>{item.fcode}H</Tag>
                                        <Text style={{ color: '#6a9955' }}>{item.message}</Text>
                                        <Text style={{ color: '#d4d4d4', marginLeft: 8, display: 'block', wordBreak: 'break-all', opacity: 0.8 }}>{item.data}</Text>
                                    </div>
                                )}
                            />
                        </div>
                    </div>
                </Content>
            </Layout>

            {/* Modals */}

            <Modal
                title="添加新站点"
                open={addDeviceModalVisible}
                onCancel={() => setAddDeviceModalVisible(false)}
                onOk={() => addDeviceForm.submit()}
            >
                <Form form={addDeviceForm} layout="vertical" onFinish={handleAddDevice}>
                    <Form.Item label="站点名称" name="name" rules={[{ required: true }]}>
                        <Input placeholder="输入站点名称" />
                    </Form.Item>
                </Form>
            </Modal>
        </Layout>
    );
}

export default App;
