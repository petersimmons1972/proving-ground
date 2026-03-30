def process_data(input_list, filter_val, transform_type, output_format):
    result = []
    temp = []
    temp2 = []
    for i in range(len(input_list)):
        item = input_list[i]
        if item is None:
            continue
        if type(item) == str:
            if len(item) == 0:
                continue
            if filter_val is not None:
                if filter_val in item:
                    temp.append(item)
                else:
                    continue
            else:
                temp.append(item)
        elif type(item) == int or type(item) == float:
            if filter_val is not None:
                if item == filter_val:
                    temp.append(item)
                else:
                    continue
            else:
                temp.append(item)
        elif type(item) == list:
            for j in range(len(item)):
                sub = item[j]
                if sub is None:
                    continue
                if filter_val is not None:
                    if sub == filter_val or (type(sub) == str and type(filter_val) == str and filter_val in sub):
                        temp.append(sub)
                else:
                    temp.append(sub)
        elif type(item) == dict:
            if "value" in item:
                v = item["value"]
                if filter_val is not None:
                    if v == filter_val or (type(v) == str and type(filter_val) == str and filter_val in v):
                        temp.append(v)
                else:
                    temp.append(v)
    if transform_type == "upper":
        for i in range(len(temp)):
            t = temp[i]
            if type(t) == str:
                temp2.append(t.upper())
            else:
                temp2.append(t)
    elif transform_type == "lower":
        for i in range(len(temp)):
            t = temp[i]
            if type(t) == str:
                temp2.append(t.lower())
            else:
                temp2.append(t)
    elif transform_type == "double":
        for i in range(len(temp)):
            t = temp[i]
            if type(t) == int or type(t) == float:
                temp2.append(t * 2)
            elif type(t) == str:
                temp2.append(t + t)
            else:
                temp2.append(t)
    else:
        temp2 = list(temp)
    if output_format == "list":
        result = temp2
    elif output_format == "set":
        result = list(set(temp2))
    elif output_format == "count":
        result = len(temp2)
    elif output_format == "first":
        result = temp2[0] if temp2 else None
    elif output_format == "last":
        result = temp2[-1] if temp2 else None
    else:
        result = temp2
    return result
