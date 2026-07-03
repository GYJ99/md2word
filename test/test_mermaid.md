# md2word Mermaid 复杂测试

本文档用于测试各种复杂 Mermaid 流程图在 Word 文档中的渲染和尺寸适配。

---

## 1. 多层级架构图（类似业务总体架构）

参照用户提供的多层架构图，包含分组、子图、跨子图连接。

```mermaid
flowchart TB
    subgraph 交互层["交互层"]
        UI1["内部门户网站<br/>100并发"]
        UI2["监控大屏<br/>10路并发"]
    end

    subgraph 业务编排层["业务编排层"]
        API1["API网关<br/>认证/限流/审计"]
        API2["低代码API"]
        API3["服务总线"]
    end

    subgraph 服务能力层["服务能力层"]
        SVC1["多源数据获取与处理<br/>接入/解析/分发"]
        SVC2["一般数据加工处理<br/>规则转换/并行/落盘"]
        SVC3["一般数据加工处理<br/>规则转换/并行/落盘"]
        SVC4["产品质量检验"]
        SVC5["运行质量与监控<br/>调度/监控/告警"]
        SVC6["综合数据库管理<br/>入库/查询/备份"]
        SVC7["用户评价与发布<br/>评价/发布/门户"]
    end

    subgraph 资源层["资源层"]
        R1["算力调度服务<br/>128核/152GB"]
        R2["存储服务集群<br/>64TB"]
        R3["存储服务节点<br/>20TB"]
    end

    UI1 --> API1
    UI1 --> API3
    UI2 --> API1
    API1 --> API2
    API2 --> SVC1
    API2 --> SVC2
    API2 --> SVC3
    API2 --> SVC4
    API2 --> SVC5
    API2 --> SVC6
    API3 --> SVC1
    API3 --> SVC7
    SVC1 --> R1
    SVC2 --> R1
    SVC3 --> R1
    SVC4 --> R1
    SVC5 --> R2
    SVC6 --> R2
    SVC7 --> R3
```

---

## 2. 纵向决策流程图（高度大于宽度）

测试高度方向的图是否会被压扁或正确缩放。

```mermaid
flowchart TD
    A[用户登录] --> B{身份验证}
    B -->|通过| C[加载权限]
    B -->|失败| D[锁定账号]
    D --> E[发送告警邮件]
    E --> F[记录审计日志]
    C --> G{权限类型}
    G -->|管理员| H[进入管理后台]
    G -->|普通用户| I[进入工作台]
    G -->|访客| J[受限浏览]
    H --> K[系统监控]
    H --> L[用户管理]
    H --> M[数据导出]
    I --> N[业务操作]
    I --> O[报表查看]
    I --> P[数据录入]
    J --> Q[只读浏览]
    K --> R[完成]
    L --> R
    M --> R
    N --> R
    O --> R
    P --> R
    Q --> R
```

---

## 3. 时序图

```mermaid
sequenceDiagram
    participant U as 用户
    participant F as 前端
    participant A as API网关
    participant S as 业务服务
    participant D as 数据库
    participant C as 缓存

    U->>F: 提交查询请求
    F->>A: HTTP POST /query
    A->>A: 鉴权 & 限流校验
    A->>S: 转发请求
    S->>C: 读取缓存
    alt 缓存命中
        C-->>S: 返回缓存数据
    else 缓存未命中
        S->>D: 查询数据库
        D-->>S: 返回结果
        S->>C: 写入缓存
    end
    S-->>A: 业务响应
    A-->>F: HTTP 200
    F-->>U: 渲染结果
```

---

## 4. 甘特图

```mermaid
gantt
    title md2word 项目迭代计划
    dateFormat YYYY-MM-DD
    section 第一阶段
    需求分析       :a1, 2026-07-01, 7d
    架构设计       :a2, after a1, 5d
    section 第二阶段
    核心转换器     :b1, after a2, 14d
    DOCX 输出      :b2, after b1, 10d
    Mermaid 渲染   :b3, after b1, 7d
    section 第三阶段
    GUI 开发       :c1, after b2, 14d
    集成测试       :c2, after b3, 7d
    发布打包       :c3, after c2, 3d
```

---

## 5. 横向超宽流程图（容易溢出页面）

测试极端宽度情况下的边界处理。

```mermaid
flowchart LR
    A[数据采集] --> B[数据清洗] --> C[数据转换] --> D[数据校验] --> E[数据存储] --> F[数据索引] --> G[数据查询] --> H[数据可视化] --> I[报表生成] --> J[导出分发]
```

---

## 6. 类图

```mermaid
classDiagram
    class Document {
        +Title string
        +Author string
        +Elements []Element
        +Save(path string) error
        +AddParagraph(p Element)
    }
    class Element {
        <<interface>>
        +ToXML() string
    }
    class Paragraph {
        +StyleID string
        +Runs []Run
        +AddRun(text string) Run
    }
    class Run {
        +Text string
        +Bold bool
        +Italic bool
        +FontSize float64
    }
    class Image {
        +RelID string
        +Width int64
        +Height int64
    }
    Document o-- Element
    Paragraph ..|> Element
    Run ..|> Element
    Image ..|> Element
```

---

## 7. 状态机图

```mermaid
stateDiagram-v2
    [*] --> 待处理
    待处理 --> 处理中: 开始处理
    处理中 --> 已完成: 处理成功
    处理中 --> 失败: 处理异常
    失败 --> 处理中: 重试
    已完成 --> [*]
    失败 --> [*]: 超过最大重试次数
```

---

## 8. 圆形节点决策图

```mermaid
flowchart TD
    Start((开始)) --> Input[/用户输入/]
    Input --> Validate{参数校验}
    Validate -->|不通过| Error[显示错误]
    Error --> Input
    Validate -->|通过| Process[处理中...]
    Process --> Decision{业务判断}
    Decision -->|分支A| PathA[[复杂处理A]]
    Decision -->|分支B| PathB((简单处理B))
    PathA --> End((结束))
    PathB --> End
```
