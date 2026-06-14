package com.generate.demo.user.impl;

import com.generate.demo.common.api.PageResult;
import com.generate.demo.common.exception.BizException;
import com.generate.demo.common.exception.ErrorCode;
import com.generate.demo.user.User;
import com.generate.demo.user.UserMapper;
import com.generate.demo.user.UserService;
import com.generate.demo.user.dto.UserCreateRequest;
import com.generate.demo.user.dto.UserResponse;
import com.generate.demo.user.dto.UserUpdateRequest;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Slf4j
@Service
@RequiredArgsConstructor
public class UserServiceImpl implements UserService {

    private final UserMapper userMapper;

    @Override
    public UserResponse getById(Long id) {
        User user = userMapper.selectById(id);
        if (user == null) {
            throw new BizException(ErrorCode.DATA_NOT_FOUND, "用户");
        }
        return toResponse(user);
    }

    @Override
    public PageResult<UserResponse> page(int page, int size) {
        Page<User> p = userMapper.selectPage(
                new Page<>(page, size),
                new LambdaQueryWrapper<User>().orderByDesc(User::getId));

        var records = p.getRecords().stream().map(this::toResponse).toList();
        return PageResult.of(p.getTotal(), (int) p.getCurrent(), (int) p.getSize(), records);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public Long create(UserCreateRequest request) {
        User user = new User();
        user.setUsername(request.getUsername());
        user.setEmail(request.getEmail());
        user.setStatus(1);
        userMapper.insert(user);
        log.info("用户创建成功 id={}, username={}", user.getId(), user.getUsername());
        return user.getId();
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void update(UserUpdateRequest request) {
        User exist = userMapper.selectById(request.getId());
        if (exist == null) {
            throw new BizException(ErrorCode.DATA_NOT_FOUND, "用户");
        }
        User user = new User();
        user.setId(request.getId());
        user.setUsername(request.getUsername());
        user.setEmail(request.getEmail());
        userMapper.updateById(user);
        log.info("用户更新成功 id={}", user.getId());
    }

    @Override
    public void delete(Long id) {
        User exist = userMapper.selectById(id);
        if (exist == null) {
            throw new BizException(ErrorCode.DATA_NOT_FOUND, "用户");
        }
        userMapper.deleteById(id);
        log.info("用户删除成功 id={}", id);
    }

    private UserResponse toResponse(User user) {
        UserResponse r = new UserResponse();
        r.setId(user.getId());
        r.setUsername(user.getUsername());
        r.setEmail(user.getEmail());
        r.setStatus(user.getStatus());
        r.setCreatedAt(user.getCreatedAt());
        r.setUpdatedAt(user.getUpdatedAt());
        return r;
    }
}
