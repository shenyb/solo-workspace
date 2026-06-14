@echo off
REM ─── gen-tables.bat ───
REM 从数据库表生成 MyBatis-Plus 代码
REM 用法: bin\gen-tables.bat

cd /d "%~dp0.."

echo === 编译 Generator ===
call mvn compile -pl {{projectName}}-dao -am -q

echo === 运行 Generator ===
call mvn exec:java -pl {{projectName}}-dao ^
    -Dexec.mainClass="{{basePackage}}.generator.Generator" ^
    -Dexec.classpathScope=compile ^
    -q

echo === 完成 ===
