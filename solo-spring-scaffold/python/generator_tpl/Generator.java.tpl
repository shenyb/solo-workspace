package {{basePackage}}.generator;

import com.baomidou.mybatisplus.generator.FastAutoGenerator;
import com.baomidou.mybatisplus.generator.config.OutputFile;
import com.baomidou.mybatisplus.generator.config.rules.DbColumnType;
import com.baomidou.mybatisplus.generator.engine.FreemarkerTemplateEngine;

import java.sql.Types;
import java.util.Collections;

/**
 * MyBatis-Plus 代码生成器
 * <p>
 * 用法：
 *   1. 配置下方数据库连接信息
 *   2. 运行 main()
 *   3. 生成的代码自动合并到 dao / service / web 模块
 */
public class Generator {

    public static void main(String[] args) {
        String url = "{{dbUrl}}";
        String username = "{{dbUser}}";
        String password = "{{dbPass}}";
        String projectPath = System.getProperty("user.dir");
        String basePackage = "{{basePackage}}";
{% if dbTables %}
        String[] tables = {
{% for table in dbTables %}
            "{{table}}",
{% endfor %}
        };
{% else %}
        // ─── 要生成的表名 ───
        String[] tables = {
            "user",
            // 添加更多表...
        };
{% endif %}

        FastAutoGenerator.create(url, username, password)
            .globalConfig(builder -> builder
                .author("solo-spring-scaffold")
                .outputDir(projectPath + "/{{projectName}}-dao/src/main/java")
                .commentDate("yyyy-MM-dd")
            )
            .packageConfig(builder -> builder
                .parent(basePackage)
                .moduleName("dao")
                .entity("entity")
                .mapper("mapper")
                .service("service")
                .serviceImpl("service.impl")
                .controller("controller")
                .pathInfo(Collections.singletonMap(OutputFile.xml, projectPath + "/{{projectName}}-dao/src/main/resources/mapper"))
            )
            .strategyConfig(builder -> builder
                .addInclude(tables)
                .addTablePrefix("t_", "sys_")

                // Entity
                .entityBuilder()
                .enableLombok()
                .enableTableFieldAnnotation()
                .logicDeleteColumnName("deleted")
                .versionColumnName("version")

                // Mapper
                .mapperBuilder()
                .enableBaseColumnList()
                .enableMapperAnnotation()

                // Service
                .serviceBuilder()
                .formatServiceFileName("%sService")
                .formatServiceImplFileName("%sServiceImpl")

                // Controller
                .controllerBuilder()
                .enableRestStyle()
                .formatFileName("%sController")
            )
            .templateEngine(new FreemarkerTemplateEngine())
            .execute();
    }
}
