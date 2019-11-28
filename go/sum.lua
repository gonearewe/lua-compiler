--chapter 13
function div0(a, b)
    if b == 0 then
      error("DIV BY ZERO !")
    else
      return a / b
    end
  end
  
function div1(a, b) return div0(a, b) end
function div2(a, b) return div1(a, b) end

ok, result = pcall(div2, 4, 2); print(ok, result)
ok, err = pcall(div2, 5, 0);    print(ok, err)
ok, err = pcall(div2, {}, {});  print(ok, err)

-- chapter 12
-- t={a=1,b=2,c=3}
-- for k,v in pairs(t) do
--     print(k,v)
-- end

-- t={"a","b","c"}
-- for k,v in ipairs(t) do
--     print(k,v)
-- end

-- chapter 10
-- function newCounter(  )
--     local count =0
--     return function ()
--         count=count+1
--         return count
--     end
-- end

-- c1=newCounter()
-- print(c1())
-- print(c1())

-- c2=newCounter()
-- print(c2())
-- print(c1())
-- print(c2())

-- chapter 9
-- print(123,{2,5,89},"Hello World")

-- chapter 8
-- local function max( ... )
--     local args={...}
--     local val,idx
--     for i=1,#args do
--         if val==nil or args[i]>val then
--             val,idx =args[i],i
--         end
--     end
--     return val,idx
-- end

-- local function assert( v )
--     if not v then fail() end
-- end

-- local v1=max(3,9,7,128,35)
-- assert(v1==128)
-- local v2,i2=max(3,9,7,128,35)
-- assert(v2==128 and i2==4)
-- local v3,i3=max(max(3,9,7,128,35))
-- assert(v3==128 and i3==1)
-- local t={max(3,9,7,128,35)}
-- assert(t[1]==128 and t[2]==4)

-- chapter 7
-- local t={"a","b","c"}
-- t[2]="8"
-- t["foo"]="Bar"
-- local s=t[3]..t[2]..t[1]..t["foo"]..#t

    -- local t={"a","b","c"}
-- t[2]="8"

-- local s=t[3]..t[2]..t[1]..t["foo"]..#t